package commands

import (
	"errors"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/path"
)

// MakeModule is `make:module <name>`, like internachi/modular's: it creates app-modules/<name> and registers the
// module in bootstrap/modules.go.
type MakeModule struct {
}

// Signature The name and signature of the console command.
func (r *MakeModule) Signature() string {
	return "make:module"
}

// Description The console command description.
func (r *MakeModule) Description() string {
	return "Create a module in app-modules/<name> and register it"
}

// Extend The console command extend.
func (r *MakeModule) Extend() command.Extend {
	return command.Extend{Category: "make"}
}

var moduleName = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// ModulePackage is the Go package name of a module folder: "role-permission" -> "rolepermission".
func ModulePackage(name string) string {
	return strings.ReplaceAll(name, "-", "")
}

// Handle Execute the console command.
func (r *MakeModule) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if !moduleName.MatchString(name) {
		err := errors.New("give a kebab-case name, e.g. make:module support-ticket")
		ctx.Error(err.Error())
		return err
	}
	if err := CreateModule(path.Base(), name); err != nil {
		ctx.Error(err.Error())
		return err
	}
	ctx.Success("module app-modules/" + name + " created and registered in bootstrap/modules.go")
	return nil
}

// CreateModule writes the module skeleton under root and registers it. It refuses to overwrite a module.
func CreateModule(root, name string) error {
	dir := filepath.Join(root, "app-modules", name)
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("app-modules/%s already exists", name)
	}
	pkg := ModulePackage(name)
	files := map[string]string{
		"module.go": fmt.Sprintf(`// Package %[1]s is the %[2]s module.
package %[1]s

import (
	"github.com/goravel/framework/contracts/route"

	"goravel/app/web"
)

// Module registers this module's routes and views; translations in lang/<locale>.json are read as
// facades.Lang(ctx).Get("%[2]s.<key>") (see app/modular for what else a module can add).
type Module struct{}

func (m *Module) Name() string      { return "%[2]s" }
func (m *Module) ViewsPath() string { return "app-modules/%[2]s/resources/views" }

func (m *Module) Routes() {
	web.Group(func(router route.Router) {
		// router.Get("/%[2]s", controller.Index)
	})
}
`, pkg, name),
		"README.md":                "# " + name + "\n\nWhat this module does, its routes and its settings.\n",
		"lang/en.json":             "{}\n",
		"lang/bn.json":             "{}\n",
		"resources/views/.gitkeep": "",
	}
	for file, content := range files {
		target := filepath.Join(dir, file)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return registerModule(root, name, pkg)
}

func registerModule(root, name, pkg string) error {
	registry := filepath.Join(root, "bootstrap", "modules.go")
	content, err := os.ReadFile(registry)
	if err != nil {
		return err
	}
	text := string(content)
	marker := "		// modules-inject (do not remove)"
	if !strings.Contains(text, marker) {
		return errors.New("bootstrap/modules.go has lost its modules-inject marker")
	}
	importLine := fmt.Sprintf("\t%s \"goravel/app-modules/%s\"\n", pkg, name)
	text = strings.Replace(text, "import (\n", "import (\n"+importLine, 1)
	text = strings.Replace(text, marker, fmt.Sprintf("\t\t&%s.Module{},\n%s", pkg, marker), 1)
	formatted, err := format.Source([]byte(text)) // keeps the imports sorted, like gofmt
	if err != nil {
		return err
	}
	return os.WriteFile(registry, formatted, 0o644)
}
