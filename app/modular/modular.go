// Package modular is the Go version of internachi/modular: every feature is a package in app-modules/<name>
// with a Module that hands its routes, migrations, commands, jobs, schedules, views and admin menu (translations: lang/<locale>.json in the module, see LangFS)
// to the app. Go has no autoloading or package discovery, so bootstrap/modules.go lists the modules
// (`go run . artisan make:module <name>` adds one there).
package modular

import (
	"sort"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/queue"
	contractsschedule "github.com/goravel/framework/contracts/schedule"
)

// Module is one feature package. It only needs a name; it adds pieces by also implementing the interfaces below.
type Module interface {
	// Name is the kebab-case folder name, e.g. "role-permission".
	Name() string
}

// HasRoutes registers routes (routes/web.go and routes/api.go for the module).
type HasRoutes interface{ Routes() }

// HasMigrations returns the module's migrations, one per table.
type HasMigrations interface {
	Migrations() []schema.Migration
}

// HasCommands returns the module's artisan commands.
type HasCommands interface {
	Commands() []console.Command
}

// HasJobs returns the module's queue jobs.
type HasJobs interface{ Jobs() []queue.Job }

// HasSchedule returns the module's scheduled tasks.
type HasSchedule interface {
	Schedule() []contractsschedule.Event
}

// HasSeeders returns the module's seeders (db:seed).
type HasSeeders interface {
	Seeders() []seeder.Seeder
}

// HasViews names the module's view folder, relative to the project root (e.g. app-modules/auth/resources/views).
// Templates in it are named "<module>::<view>" with {{ define "auth::login" }}, like Laravel's view namespaces.
type HasViews interface{ ViewsPath() string }

// HasAdminMenu adds links to the admin sidebar.
type HasAdminMenu interface {
	AdminMenu() []MenuItem
}

// MenuItem is one admin sidebar link. It is shown only to people who hold Permission.
type MenuItem struct {
	Section    string // "Overview", "Content", "People and access", "System"
	Label      string
	Href       string
	Permission string
	Icon       string // Font Awesome classes
	Order      int
}

// MenuSections is the order of the sidebar sections; unknown sections come last.
var MenuSections = []string{"Overview", "Content", "People and access", "System"}

var registered []Module

// Register keeps the module list (bootstrap/modules.go) for modules that need it at run time, such as the admin
// sidebar. Modules cannot import bootstrap, which imports them.
func Register(modules []Module) { registered = modules }

// Registered is the module list given to Register.
func Registered() []Module { return registered }

// Collect gathers one kind of item from every module that provides it.
func Collect[T any, I any](modules []Module, pick func(I) []T) []T {
	var items []T
	for _, module := range modules {
		if provider, ok := module.(I); ok {
			items = append(items, pick(provider)...)
		}
	}
	return items
}

// Menu is every module's sidebar links, sorted by section, then order, then label.
func Menu(modules []Module) []MenuItem {
	items := Collect(modules, func(m HasAdminMenu) []MenuItem { return m.AdminMenu() })
	rank := func(section string) int {
		for index, name := range MenuSections {
			if name == section {
				return index
			}
		}
		return len(MenuSections)
	}
	sort.SliceStable(items, func(a, b int) bool {
		if rank(items[a].Section) != rank(items[b].Section) {
			return rank(items[a].Section) < rank(items[b].Section)
		}
		if items[a].Order != items[b].Order {
			return items[a].Order < items[b].Order
		}
		return items[a].Label < items[b].Label
	})
	return items
}
