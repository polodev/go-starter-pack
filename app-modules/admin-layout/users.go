package adminlayout

import (
	"net/mail"
	"sort"
	"strconv"
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app-modules/auth"
	rolepermission "goravel/app-modules/role-permission"
	rpmodels "goravel/app-modules/role-permission/models"
	"goravel/app/activity"
	"goravel/app/facades"
	"goravel/app/models"
)

const usersPerPage = 25

// UsersIndex is GET /admin/users: ID first (a link), search with ?q=.
func UsersIndex(ctx http.Context) http.Response {
	query := facades.Orm().Query().Model(&models.User{}).OrderBy("id", "desc")
	search := strings.TrimSpace(ctx.Request().Query("q"))
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("(name LIKE ? OR email LIKE ?)", like, like)
	}
	var users []models.User
	var total int64
	page := PageParam(ctx)
	_ = query.Paginate(page, usersPerPage, &users, &total)
	roles := map[uint]string{}
	for _, user := range users {
		roles[user.ID] = rolepermission.RoleSlug(user.ID)
	}
	pages := int((total + usersPerPage - 1) / usersPerPage)
	return View(ctx, "admin::users-index", "Users", []Crumb{{Label: "Users"}}, map[string]any{
		"users": users, "roles": roles, "q": search, "total": total, "page": page, "pages": pages,
		"prev": page - 1, "next": nextPage(page, pages),
	})
}

func nextPage(page, pages int) int {
	if page < pages {
		return page + 1
	}
	return 0
}

func findUser(ctx http.Context) (models.User, bool) {
	var user models.User
	id := IDParam(ctx)
	if id == 0 || facades.Orm().Query().Find(&user, id) != nil || user.ID == 0 {
		return user, false
	}
	return user, true
}

func exceptionKeys(userID uint) []string {
	var keys []string
	_ = facades.Orm().Query().Model(&rpmodels.Permission{}).
		Join("join permission_user on permission_user.permission_id = permissions.id").
		Where("permission_user.user_id", userID).Pluck("permissions.key", &keys)
	sort.Strings(keys)
	return keys
}

// UserShow is GET /admin/users/{id}: read-only, with an Edit button for people with users.manage.
func UserShow(ctx http.Context) http.Response {
	user, ok := findUser(ctx)
	if !ok {
		return NotFound(ctx)
	}
	return View(ctx, "admin::user-show", user.Name, []Crumb{{Label: "Users", Href: "/admin/users"}, {Label: user.Name}},
		map[string]any{"account": user, "role": rolepermission.RoleSlug(user.ID), "exceptions": exceptionKeys(user.ID)})
}

func editData(ctx http.Context, user models.User, values map[string]any) map[string]any {
	var roles []rpmodels.Role
	_ = facades.Orm().Query().Where("is_active", true).OrderBy("rank").Find(&roles)
	var permissions []rpmodels.Permission
	_ = facades.Orm().Query().Where("is_active", true).OrderBy("id").Find(&permissions)
	assignable := []rpmodels.Permission{}
	for _, permission := range permissions {
		if !strings.HasPrefix(permission.Key, "roles.") {
			assignable = append(assignable, permission)
		}
	}
	selected := map[string]bool{}
	for _, key := range exceptionKeys(user.ID) {
		selected[key] = true
	}
	data := map[string]any{
		"account": user, "name": user.Name, "email": user.Email, "roleList": roles, "role": rolepermission.RoleSlug(user.ID),
		"permissionList": assignable, "selected": selected,
	}
	for key, value := range values {
		data[key] = value
	}
	return data
}

// UserEdit is GET /admin/users/{id}/edit.
func UserEdit(ctx http.Context) http.Response {
	user, ok := findUser(ctx)
	if !ok {
		return NotFound(ctx)
	}
	return View(ctx, "admin::user-edit", "Edit "+user.Name, editCrumbs(user), editData(ctx, user, nil))
}

func editCrumbs(user models.User) []Crumb {
	return []Crumb{{Label: "Users", Href: "/admin/users"}, {Label: user.Name, Href: "/admin/users/" + strconv.Itoa(int(user.ID))}, {Label: "Edit"}}
}

// UserUpdate is POST /admin/users/{id}: name and email (users.manage); role and exceptions only with roles.assign.
func UserUpdate(ctx http.Context) http.Response {
	user, ok := findUser(ctx)
	if !ok {
		return NotFound(ctx)
	}
	name := strings.TrimSpace(ctx.Request().Input("name"))
	email := auth.NormaliseEmail(ctx.Request().Input("email"))
	errors := map[string]string{}
	if name == "" || len(name) > 255 {
		errors["name"] = "Enter a name (up to 255 characters)."
	}
	if _, err := mail.ParseAddress(email); err != nil || len(email) > 255 {
		errors["email"] = "Enter a valid email address."
	} else if email != user.Email {
		if taken, _ := facades.Orm().Query().Model(&models.User{}).Where("email", email).Exists(); taken {
			errors["email"] = "Another account already uses this email."
		}
	}
	canAssign := rolepermission.Can(ctx)("roles.assign")
	role := ctx.Request().Input("role")
	exceptions := ctx.Request().InputArray("exceptions")
	if canAssign && rolepermission.RoleSlug(user.ID) == "developer" && role != "developer" && developerCount() <= 1 {
		errors["role"] = "Keep at least one developer: this is the last one."
	}
	if len(errors) > 0 {
		return ViewStatus(ctx, http.StatusUnprocessableEntity, "admin::user-edit", "Edit "+user.Name, editCrumbs(user),
			editData(ctx, user, map[string]any{"name": name, "email": email, "errors": errors}))
	}

	before := map[string]any{"name": user.Name, "email": user.Email}
	user.Name, user.Email = name, email
	if err := facades.Orm().Query().Save(&user); err != nil {
		return ctx.Response().Redirect(http.StatusSeeOther, "/admin/users/"+strconv.Itoa(int(user.ID))+"/edit?error=save")
	}
	if canAssign {
		if err := rolepermission.AssignRole(user.ID, role); err != nil {
			return ctx.Response().Redirect(http.StatusSeeOther, "/admin/users/"+strconv.Itoa(int(user.ID))+"/edit?error=role")
		}
		if err := rolepermission.SetExceptions(user.ID, exceptions); err != nil {
			return ctx.Response().Redirect(http.StatusSeeOther, "/admin/users/"+strconv.Itoa(int(user.ID))+"/edit?error=exceptions")
		}
	}
	if editor := auth.CurrentUser(ctx); editor != nil {
		entry := activity.In("users").CausedBy(editor.ID).PerformedOn("user", user.ID).Event("updated").
			Changes(before, map[string]any{"name": user.Name, "email": user.Email})
		if canAssign {
			entry.With("role", role).With("exceptions", exceptions)
		}
		entry.LogQuietly("updated user")
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/admin/users/"+strconv.Itoa(int(user.ID))+"?saved=user")
}

func developerCount() int64 {
	count, _ := facades.DB().Table("role_user").Join("join roles on roles.id = role_user.role_id").
		Where("roles.slug", "developer").Count()
	return count
}
