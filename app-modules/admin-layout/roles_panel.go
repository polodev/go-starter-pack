package adminlayout

import (
	"sort"

	"github.com/goravel/framework/contracts/http"

	"goravel/app-modules/auth"
	rolepermission "goravel/app-modules/role-permission"
	rpmodels "goravel/app-modules/role-permission/models"
	"goravel/app/activity"
	"goravel/app/facades"
	"goravel/app/models"
)

// The roles and permissions panel follows the Laravel apps' six parts: Overview, Roles, Permissions,
// Permission Matrix, User Assignments and Backend Links. Roles are edited in code and synced; people get their
// role on Users > Edit.

var panelCrumb = Crumb{Label: "Roles & permissions", Href: "/admin/role-permissions"}

func activeRoles() []rpmodels.Role {
	var roles []rpmodels.Role
	_ = facades.Orm().Query().Where("is_active", true).OrderBy("rank").Find(&roles)
	return roles
}

func activePermissions() []rpmodels.Permission {
	var permissions []rpmodels.Permission
	_ = facades.Orm().Query().Where("is_active", true).Find(&permissions)
	sort.Slice(permissions, func(a, b int) bool {
		if permissions[a].Group != permissions[b].Group {
			return permissions[a].Group < permissions[b].Group
		}
		return permissions[a].Key < permissions[b].Key
	})
	return permissions
}

// rolePermissionKeys maps role slug -> set of its permission keys.
func rolePermissionKeys() map[string]map[string]bool {
	type row struct {
		Slug          string
		PermissionKey string
	}
	var rows []row
	_ = facades.DB().Table("permission_role").
		Join("join roles on roles.id = permission_role.role_id").
		Join("join permissions on permissions.id = permission_role.permission_id").
		Select("roles.slug as slug", "permissions.key as permission_key").Get(&rows)
	keys := map[string]map[string]bool{}
	for _, item := range rows {
		if keys[item.Slug] == nil {
			keys[item.Slug] = map[string]bool{}
		}
		keys[item.Slug][item.PermissionKey] = true
	}
	return keys
}

// RolesOverview is GET /admin/role-permissions.
func RolesOverview(ctx http.Context) http.Response {
	assigned, _ := facades.DB().Table("role_user").Count()
	exceptions, _ := facades.DB().Table("permission_user").Count()
	return View(ctx, "admin::roles-overview", "Roles & permissions", []Crumb{{Label: "Roles & permissions"}}, map[string]any{
		"tab": "overview", "roleCount": len(activeRoles()), "permissionCount": len(activePermissions()),
		"assigned": assigned, "exceptions": exceptions, "unmapped": rolepermission.Unmapped(),
	})
}

// RolesIndex is GET /admin/role-permissions/roles.
func RolesIndex(ctx http.Context) http.Response {
	keys := rolePermissionKeys()
	counts := map[string]int{}
	for slug, set := range keys {
		counts[slug] = len(set)
	}
	return View(ctx, "admin::roles-index", "Roles", []Crumb{panelCrumb, {Label: "Roles"}},
		map[string]any{"tab": "roles", "roles": activeRoles(), "counts": counts})
}

// RoleShow is GET /admin/role-permissions/roles/{slug}.
func RoleShow(ctx http.Context) http.Response {
	var role rpmodels.Role
	slug := ctx.Request().Route("slug")
	if facades.Orm().Query().Where("slug", slug).First(&role) != nil || role.ID == 0 {
		return NotFound(ctx)
	}
	granted := rolePermissionKeys()[slug]
	var permissions []rpmodels.Permission
	for _, permission := range activePermissions() {
		if granted[permission.Key] {
			permissions = append(permissions, permission)
		}
	}
	var users []models.User
	_ = facades.Orm().Query().Model(&models.User{}).Join("join role_user on role_user.user_id = users.id").
		Where("role_user.role_id", role.ID).OrderBy("users.id").Find(&users)
	return View(ctx, "admin::role-show", role.Name, []Crumb{panelCrumb, {Label: "Roles", Href: "/admin/role-permissions/roles"}, {Label: role.Name}},
		map[string]any{"tab": "roles", "role": role, "permissions": permissions, "users": users})
}

// PermissionsIndex is GET /admin/role-permissions/permissions.
func PermissionsIndex(ctx http.Context) http.Response {
	return View(ctx, "admin::permissions-index", "Permissions", []Crumb{panelCrumb, {Label: "Permissions"}},
		map[string]any{"tab": "permissions", "permissions": activePermissions()})
}

// Matrix is GET /admin/role-permissions/matrix: every permission against every role.
func Matrix(ctx http.Context) http.Response {
	return View(ctx, "admin::roles-matrix", "Permission matrix", []Crumb{panelCrumb, {Label: "Permission matrix"}},
		map[string]any{"tab": "matrix", "roles": activeRoles(), "permissions": activePermissions(), "keys": rolePermissionKeys()})
}

// Assignments is GET /admin/role-permissions/users: who has which role, and direct exceptions.
func Assignments(ctx http.Context) http.Response {
	type row struct {
		ID    uint
		Name  string
		Email string
		Role  string
	}
	var rows []row
	_ = facades.DB().Table("users").Join("join role_user on role_user.user_id = users.id").
		Join("join roles on roles.id = role_user.role_id").
		Select("users.id as id", "users.name as name", "users.email as email", "roles.name as role").
		OrderBy("roles.rank", "desc").Get(&rows)
	exceptions := map[uint][]string{}
	for _, item := range rows {
		exceptions[item.ID] = exceptionKeys(item.ID)
	}
	return View(ctx, "admin::roles-assignments", "User assignments", []Crumb{panelCrumb, {Label: "User assignments"}},
		map[string]any{"tab": "users", "rows": rows, "exceptions": exceptions})
}

// BackendLinks is GET /admin/role-permissions/links: every admin route and the permission it needs.
func BackendLinks(ctx http.Context) http.Response {
	return View(ctx, "admin::roles-links", "Backend links", []Crumb{panelCrumb, {Label: "Backend links"}},
		map[string]any{"tab": "links", "routes": rolepermission.Declared(), "unmapped": rolepermission.Unmapped()})
}

// SyncRoles is POST /admin/role-permissions/sync (roles.sync): the same as role-permission:sync.
func SyncRoles(ctx http.Context) http.Response {
	result, err := rolepermission.Sync()
	if err != nil {
		return ctx.Response().Redirect(http.StatusSeeOther, "/admin/role-permissions?error=sync")
	}
	if user := auth.CurrentUser(ctx); user != nil {
		activity.In("roles").CausedBy(user.ID).Event("synced").With("permissions", result.Permissions).
			With("roles", result.Roles).With("switched_off", result.Deactivated).LogQuietly("synced roles and permissions")
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/admin/role-permissions?saved=sync")
}
