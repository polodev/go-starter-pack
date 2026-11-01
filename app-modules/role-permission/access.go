package rolepermission

import (
	"strconv"
	"strings"
	"sync"

	"goravel/app-modules/role-permission/models"
	"goravel/app/facades"
)

// The person's granted keys are cached for a short while per user; any change to roles or assignments clears it.
var (
	cacheMutex sync.Mutex
	cache      = map[uint][]string{}
)

func forgetAll() {
	cacheMutex.Lock()
	cache = map[uint][]string{}
	cacheMutex.Unlock()
}

// Forget clears one person's cached permissions (after assigning a role or an exception).
func Forget(userID uint) {
	cacheMutex.Lock()
	delete(cache, userID)
	cacheMutex.Unlock()
}

// SuperUserIDs come from ROLE_PERMISSION_SUPER_USER_IDS (comma separated): they pass every check, like
// ADVANCED_AUTH_SUPER_USER_IDS in the Laravel apps. Keep it empty unless you need a way back in.
func SuperUserIDs() map[uint]bool {
	ids := map[uint]bool{}
	text, _ := facades.Config().Env("ROLE_PERMISSION_SUPER_USER_IDS", "").(string)
	for _, part := range strings.Split(text, ",") {
		if id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64); err == nil && id > 0 {
			ids[uint(id)] = true
		}
	}
	return ids
}

// GrantedKeys is every key the person holds: from their active role's active permissions plus their active
// direct exceptions. A developer holds "*".
func GrantedKeys(userID uint) []string {
	cacheMutex.Lock()
	if keys, ok := cache[userID]; ok {
		cacheMutex.Unlock()
		return keys
	}
	cacheMutex.Unlock()

	var keys []string
	if RoleSlug(userID) == "developer" {
		keys = []string{"*"}
	} else {
		_ = facades.Orm().Query().Model(&models.Permission{}).
			Join("join permission_role on permission_role.permission_id = permissions.id").
			Join("join role_user on role_user.role_id = permission_role.role_id").
			Join("join roles on roles.id = role_user.role_id").
			Where("role_user.user_id", userID).Where("roles.is_active", true).Where("permissions.is_active", true).
			Pluck("permissions.key", &keys)
		var direct []string
		_ = facades.Orm().Query().Model(&models.Permission{}).
			Join("join permission_user on permission_user.permission_id = permissions.id").
			Where("permission_user.user_id", userID).Where("permissions.is_active", true).
			Pluck("permissions.key", &direct)
		keys = append(keys, direct...)
	}
	cacheMutex.Lock()
	cache[userID] = keys
	cacheMutex.Unlock()
	return keys
}

// RoleSlug is the person's active role, or "".
func RoleSlug(userID uint) string {
	var slugs []string
	_ = facades.Orm().Query().Model(&models.Role{}).
		Join("join role_user on role_user.role_id = roles.id").
		Where("role_user.user_id", userID).Where("roles.is_active", true).Pluck("roles.slug", &slugs)
	if len(slugs) == 0 {
		return ""
	}
	return slugs[0]
}

// HasPermission is the one check everything uses (middleware, the Gate, templates).
func HasPermission(userID uint, key string) bool {
	if userID == 0 {
		return false
	}
	if SuperUserIDs()[userID] {
		return true
	}
	granted := GrantedKeys(userID)
	// Anyone with a role or an exception may open the backend (the dashboard), like the Laravel apps.
	if key == "dashboard.view" && len(granted) > 0 {
		return true
	}
	return grants(granted, key)
}

// IsStaff reports whether the person has any role or exception (may see the admin area at all).
func IsStaff(userID uint) bool { return HasPermission(userID, "dashboard.view") }
