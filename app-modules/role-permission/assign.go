package rolepermission

import (
	"errors"
	"strings"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app-modules/role-permission/models"
	"goravel/app/facades"
)

// AssignRole gives a person exactly one role (a person has one platform role, like the Laravel apps).
// An empty slug removes the role.
func AssignRole(userID uint, slug string) error {
	var role models.Role
	if slug != "" {
		if err := facades.Orm().Query().Where("slug", slug).Where("is_active", true).First(&role); err != nil || role.ID == 0 {
			return errors.New("no active role " + slug + "; run role-permission:sync")
		}
	}
	err := facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Where("user_id", userID).Delete(&models.RoleUser{}); err != nil {
			return err
		}
		if role.ID == 0 {
			return nil
		}
		return tx.Create(&models.RoleUser{UserID: userID, RoleID: role.ID})
	})
	Forget(userID)
	return err
}

// SetExceptions replaces a person's direct permissions. Role-management permissions (roles.*) can never be
// granted this way: they come only with the developer role.
func SetExceptions(userID uint, keys []string) error {
	ids := []uint{}
	for _, key := range keys {
		if strings.HasPrefix(key, "roles.") {
			return errors.New(key + " can only come with a role, not as an exception")
		}
		var permission models.Permission
		if err := facades.Orm().Query().Where(&models.Permission{Key: key}).Where("is_active", true).First(&permission); err != nil || permission.ID == 0 {
			return errors.New("no active permission " + key)
		}
		ids = append(ids, permission.ID)
	}
	err := facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Where("user_id", userID).Delete(&models.PermissionUser{}); err != nil {
			return err
		}
		for _, id := range ids {
			if err := tx.Create(&models.PermissionUser{UserID: userID, PermissionID: id}); err != nil {
				return err
			}
		}
		return nil
	})
	Forget(userID)
	return err
}
