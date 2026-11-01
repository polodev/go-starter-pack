package rolepermission

import (
	"sort"
	"strings"
	"sync"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"

	"goravel/app-modules/auth"
	"goravel/app/web"
)

// Backend routes fail closed: every /admin route is declared with its permission, and a route that has none is
// refused even for a developer (the Laravel apps' AdvancedRouteAuthorizationMiddleware).
var (
	registryMutex sync.RWMutex
	registry      = map[string]string{} // "GET /admin/users/{id}" -> "users.view"
)

// AdminRouter declares /admin routes with their permission.
type AdminRouter struct{ router route.Router }

func (a *AdminRouter) add(method, path, permission string) string {
	full := "/admin" + strings.TrimSuffix(path, "/")
	registryMutex.Lock()
	registry[method+" "+full] = permission
	registryMutex.Unlock()
	if path == "/" {
		return "" // the group's own address, /admin, not /admin/ (which would redirect)
	}
	return path
}

// Get declares GET /admin<path> needing permission.
func (a *AdminRouter) Get(path, permission string, handler http.HandlerFunc) {
	a.router.Get(a.add("GET", path, permission), handler)
}

// Post declares POST /admin<path> needing permission.
func (a *AdminRouter) Post(path, permission string, handler http.HandlerFunc) {
	a.router.Post(a.add("POST", path, permission), handler)
}

// Admin registers backend routes: signed in, verified email, and the declared permission for each route.
func Admin(register func(admin *AdminRouter)) {
	web.Group(func(router route.Router) {
		router.Prefix("/admin").Middleware(&auth.Authenticate{}, &auth.Verified{}, &Authorize{}).Group(func(group route.Router) {
			register(&AdminRouter{router: group})
		})
	})
}

// PermissionFor is the permission declared for a route ("" when the route was never declared).
func PermissionFor(method, pattern string) string {
	if method == "HEAD" {
		method = "GET"
	}
	pattern = "/" + strings.Trim(pattern, "/")
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	return registry[method+" "+pattern]
}

// Declared lists every declared admin route and its permission, sorted (for the Backend Links screen).
func Declared() [][2]string {
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	rows := make([][2]string, 0, len(registry))
	for route, permission := range registry {
		rows = append(rows, [2]string{route, permission})
	}
	sort.Slice(rows, func(a, b int) bool { return rows[a][0] < rows[b][0] })
	return rows
}

// Authorize checks the declared permission of the matched admin route.
type Authorize struct{}

func (r *Authorize) Signature() string { return "role-permission:authorize" }

func (r *Authorize) Handle(ctx http.Context) {
	permission := PermissionFor(ctx.Request().Method(), ctx.Request().OriginPath())
	user := auth.CurrentUser(ctx)
	if permission == "" || user == nil || !HasPermission(user.ID, permission) {
		if page, ok := web.ViewStatus(ctx, http.StatusForbidden, "site::error", map[string]any{
			"code": http.StatusForbidden, "message": web.T(ctx, "site-layout.forbidden"), "title": web.T(ctx, "site-layout.forbidden"),
		}).(http.AbortableResponse); ok {
			_ = page.Abort()
		}
		return
	}
	ctx.Request().Next()
}

// Can returns the person's check for templates: {{ if call .can "users.view" }}.
func Can(ctx http.Context) func(key string) bool {
	return func(key string) bool {
		user := auth.CurrentUser(ctx)
		return user != nil && HasPermission(user.ID, key)
	}
}
