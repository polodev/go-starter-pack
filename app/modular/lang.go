package modular

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/support/path"
)

// LangFS lets Goravel's translator (facades.Lang) read every module's own translations, like Laravel's
// loadJsonTranslationsFrom. Goravel asks for "<locale>/<group>.json"; the group is the module name, so
// facades.Lang(ctx).Get("site-layout.nav-about") reads app-modules/site-layout/lang/<locale>.json, key "nav-about".
// Anything else falls back to the app's lang folder (lang/<locale>/<group>.json, lang/<locale>.json).
type LangFS struct{}

func (LangFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	if locale, file, found := strings.Cut(name, "/"); found && !strings.Contains(file, "/") {
		group := strings.TrimSuffix(file, ".json")
		moduleFile := path.Base("app-modules", group, "lang", locale+".json")
		if opened, err := os.Open(moduleFile); err == nil {
			return opened, nil
		}
	}
	return os.Open(path.Base(filepath.Join("lang", filepath.FromSlash(name))))
}
