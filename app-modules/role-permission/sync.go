package rolepermission

import (
	"fmt"
	"sort"
	"strings"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app-modules/role-permission/models"
	"goravel/app/facades"
)

// SyncResult says what role-permission:sync changed.
type SyncResult struct {
	Permissions, Roles, Deactivated int
}

// Validate checks the definitions: every granted key exists (or is a known "area.*"), ranks rise, and each role
// includes everything of the role below it.
func Validate() error {
	known := map[string]bool{}
	for _, key := range AllPermissionKeys() {
		known[key] = true
	}
	for index, role := range Roles {
		for _, granted := range role.Permissions {
			if granted == "*" {
				continue
			}
			if area, found := strings.CutSuffix(granted, ".*"); found {
				if _, ok := Permissions[area]; !ok {
					return fmt.Errorf("role %s grants %s, but there is no area %q", role.Slug, granted, area)
				}
				continue
			}
			if !known[granted] {
				return fmt.Errorf("role %s grants %s, which is not defined", role.Slug, granted)
			}
		}
		if index > 0 {
			below := Roles[index-1]
			if role.Rank <= below.Rank {
				return fmt.Errorf("role %s must rank above %s", role.Slug, below.Slug)
			}
			for _, granted := range below.Permissions {
				if !grants(role.Permissions, granted) {
					return fmt.Errorf("role %s must include %s from %s", role.Slug, granted, below.Slug)
				}
			}
		}
	}
	return nil
}

// Sync copies the definitions into the tables in one transaction. Removed definitions are switched off, not
// deleted, so assignments survive a mistake.
func Sync() (SyncResult, error) {
	if err := Validate(); err != nil {
		return SyncResult{}, err
	}
	var result SyncResult
	err := facades.Orm().Transaction(func(tx orm.Query) error {
		keys := AllPermissionKeys()
		sort.Strings(keys)
		permissionIDs := map[string]uint{}
		for area, group := range Permissions {
			for key, description := range group {
				var permission models.Permission
				if err := tx.UpdateOrCreate(&permission, models.Permission{Key: key},
					map[string]any{"name": description, "group": area, "is_active": true}); err != nil {
					return err
				}
				permissionIDs[key] = permission.ID
				result.Permissions++
			}
		}
		var stored []models.Permission
		if err := tx.Find(&stored); err != nil {
			return err
		}
		for _, permission := range stored {
			if _, defined := permissionIDs[permission.Key]; !defined && permission.IsActive {
				if _, err := tx.Model(&models.Permission{}).Where("id", permission.ID).Update("is_active", false); err != nil {
					return err
				}
				result.Deactivated++
			}
		}

		roleSlugs := map[string]bool{}
		for _, definition := range Roles {
			roleSlugs[definition.Slug] = true
			var role models.Role
			if err := tx.UpdateOrCreate(&role, models.Role{Slug: definition.Slug}, map[string]any{
				"name": definition.Name, "rank": definition.Rank, "description": definition.Description, "is_active": true,
			}); err != nil {
				return err
			}
			if _, err := tx.Where("role_id", role.ID).Delete(&models.PermissionRole{}); err != nil {
				return err
			}
			for _, key := range keys {
				if grants(definition.Permissions, key) {
					if err := tx.Create(&models.PermissionRole{RoleID: role.ID, PermissionID: permissionIDs[key]}); err != nil {
						return err
					}
				}
			}
			result.Roles++
		}
		var storedRoles []models.Role
		if err := tx.Find(&storedRoles); err != nil {
			return err
		}
		for _, role := range storedRoles {
			if !roleSlugs[role.Slug] && role.IsActive {
				if _, err := tx.Model(&models.Role{}).Where("id", role.ID).Update("is_active", false); err != nil {
					return err
				}
				result.Deactivated++
			}
		}
		return nil
	})
	forgetAll()
	return result, err
}

// grants reports whether a list of granted patterns covers key ("*", the key itself, or "area.*").
func grants(granted []string, key string) bool {
	for _, pattern := range granted {
		if pattern == "*" || pattern == key ||
			(strings.HasSuffix(pattern, ".*") && strings.HasPrefix(key, strings.TrimSuffix(pattern, "*"))) {
			return true
		}
	}
	return false
}
