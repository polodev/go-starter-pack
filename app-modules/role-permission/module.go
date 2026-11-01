// Package rolepermission is staff authorization, the model of the Laravel apps' role-permission module:
// permissions and five ranked roles defined in code (definitions.go), copied into the tables by
// role-permission:sync; one role per person plus direct exceptions; admin routes declared with their permission
// and refused when they have none; Goravel's Gate answers through the same check.
package rolepermission

import (
	"context"

	"github.com/goravel/framework/auth/access"
	contractsaccess "github.com/goravel/framework/contracts/auth/access"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/http"

	"goravel/app-modules/auth"
	"goravel/app-modules/role-permission/migrations"
	"goravel/app/facades"
	"goravel/app/web"
)

// Module registers the tables, the commands, the Gate hook and the template helpers.
type Module struct{}

func (m *Module) Name() string { return "role-permission" }

func (m *Module) Migrations() []schema.Migration {
	return []schema.Migration{&migrations.M20261101150000CreateRolePermissionTables{}}
}

func (m *Module) Commands() []console.Command {
	return []console.Command{&SyncCommand{}, &AssignCommand{}, &AuditCommand{}}
}

func (m *Module) Routes() {
	// facades.Gate().WithContext(ctx).Allows("users.view") is answered by the role model (Laravel's Gate::before).
	facades.Gate().Before(func(ctx context.Context, ability string, arguments map[string]any) contractsaccess.Response {
		if httpContext, ok := ctx.(http.Context); ok {
			if user := auth.CurrentUser(httpContext); user != nil && HasPermission(user.ID, ability) {
				return access.NewAllowResponse()
			}
		}
		return nil
	})
	web.Compose(func(ctx http.Context, data map[string]any) {
		data["can"] = Can(ctx)
		if user := auth.CurrentUser(ctx); user != nil {
			data["canAdmin"] = IsStaff(user.ID)
		}
	})
}
