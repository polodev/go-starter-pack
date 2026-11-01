// Package models holds the role and permission models.
package models

import "github.com/goravel/framework/database/orm"

// Role is a staff role (new_employee ... developer).
type Role struct {
	orm.Model
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Rank        int     `json:"rank"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}

// Permission is one "<area>.<action>" key.
type Permission struct {
	orm.Model
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	Group       string  `json:"group"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}

// PermissionRole links a role to a permission.
type PermissionRole struct {
	RoleID       uint `gorm:"primaryKey" json:"role_id"`
	PermissionID uint `gorm:"primaryKey" json:"permission_id"`
	orm.Timestamps
}

func (PermissionRole) TableName() string { return "permission_role" }

// RoleUser gives a person a role.
type RoleUser struct {
	UserID uint `gorm:"primaryKey" json:"user_id"`
	RoleID uint `gorm:"primaryKey" json:"role_id"`
	orm.Timestamps
}

func (RoleUser) TableName() string { return "role_user" }

// PermissionUser is a direct exception: one permission for one person, outside their role.
type PermissionUser struct {
	UserID       uint `gorm:"primaryKey" json:"user_id"`
	PermissionID uint `gorm:"primaryKey" json:"permission_id"`
	orm.Timestamps
}

func (PermissionUser) TableName() string { return "permission_user" }
