package rolepermission

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	"goravel/app/facades"
	"goravel/app/models"
)

// SyncCommand is `role-permission:sync`.
type SyncCommand struct{}

func (r *SyncCommand) Signature() string { return "role-permission:sync" }
func (r *SyncCommand) Description() string {
	return "Copy the roles and permissions defined in code into the database"
}
func (r *SyncCommand) Extend() command.Extend { return command.Extend{Category: "role-permission"} }

func (r *SyncCommand) Handle(ctx console.Context) error {
	result, err := Sync()
	if err != nil {
		ctx.Error(err.Error())
		return err
	}
	ctx.Success(fmt.Sprintf("%d permissions, %d roles synced; %d switched off", result.Permissions, result.Roles, result.Deactivated))
	return nil
}

// AssignCommand is `role-permission:assign --email=you@example.com --role=developer`: how the first developer
// gets in (after that, a developer assigns roles in the admin).
type AssignCommand struct{}

func (r *AssignCommand) Signature() string   { return "role-permission:assign" }
func (r *AssignCommand) Description() string { return "Give a person a role (empty --role removes it)" }
func (r *AssignCommand) Extend() command.Extend {
	return command.Extend{Category: "role-permission", Flags: []command.Flag{
		&command.StringFlag{Name: "email", Usage: "the person's email"},
		&command.StringFlag{Name: "role", Usage: "new_employee, business_development, employee, admin or developer"},
	}}
}

func (r *AssignCommand) Handle(ctx console.Context) error {
	var user models.User
	email := strings.ToLower(strings.TrimSpace(ctx.Option("email")))
	if err := facades.Orm().Query().Where("email", email).First(&user); err != nil || user.ID == 0 {
		err := errors.New("no account with email " + email)
		ctx.Error(err.Error())
		return err
	}
	if err := AssignRole(user.ID, ctx.Option("role")); err != nil {
		ctx.Error(err.Error())
		return err
	}
	ctx.Success(email + " now has role " + fmt.Sprintf("%q", ctx.Option("role")))
	return nil
}

// AuditCommand is `role-permission:audit`: every /admin route must be declared with a permission.
type AuditCommand struct{}

func (r *AuditCommand) Signature() string { return "role-permission:audit" }
func (r *AuditCommand) Description() string {
	return "Fail when an /admin route has no permission (it would answer 403 to everyone)"
}
func (r *AuditCommand) Extend() command.Extend { return command.Extend{Category: "role-permission"} }

func (r *AuditCommand) Handle(ctx console.Context) error {
	unmapped := Unmapped()
	for _, route := range unmapped {
		ctx.Error("no permission: " + route)
	}
	if len(unmapped) > 0 {
		return fmt.Errorf("%d admin route(s) without a permission", len(unmapped))
	}
	ctx.Success(fmt.Sprintf("all %d admin routes have a permission", len(Declared())))
	return nil
}

// Unmapped lists /admin routes the router knows but no module declared.
func Unmapped() []string {
	var missing []string
	for _, info := range facades.Route().GetRoutes() {
		path := "/" + strings.Trim(info.Path, "/")
		if !strings.HasPrefix(path, "/admin") {
			continue
		}
		for _, method := range strings.Split(info.Method, "|") {
			if method == "HEAD" {
				continue
			}
			if PermissionFor(method, path) == "" {
				missing = append(missing, method+" "+path)
			}
		}
	}
	return missing
}
