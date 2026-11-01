package adminlayout

import (
	"github.com/goravel/framework/contracts/http"

	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/facades"
	"goravel/app/models"
)

// Dashboard is GET /admin: the numbers that matter at a glance.
func Dashboard(ctx http.Context) http.Response {
	users, _ := facades.Orm().Query().Model(&models.User{}).Count()
	verified, _ := facades.Orm().Query().Model(&models.User{}).WhereNotNull("email_verified_at").Count()
	staff, _ := facades.DB().Table("role_user").Count()
	waiting, _ := facades.DB().Table("jobs").Count()
	failed, _ := facades.DB().Table("failed_jobs").Count()
	// A card links to its screen only for people who may open that screen.
	can := rolepermission.Can(ctx)
	link := func(href, permission string) string {
		if can(permission) {
			return href
		}
		return ""
	}
	return View(ctx, "admin::dashboard", "Dashboard", nil, map[string]any{
		"stats": []map[string]any{
			{"label": "Users", "value": users, "href": link("/admin/users", "users.view"), "testid": "stat-users"},
			{"label": "Verified emails", "value": verified, "href": link("/admin/users", "users.view"), "testid": "stat-verified"},
			{"label": "Staff with a role", "value": staff, "href": link("/admin/role-permissions/users", "roles.view"), "testid": "stat-staff"},
			{"label": "Jobs waiting", "value": waiting, "href": "", "testid": "stat-jobs"},
			{"label": "Failed jobs", "value": failed, "href": "", "testid": "stat-failed-jobs"},
		},
	})
}
