// Package models holds the link between outside accounts and users.
package models

import "github.com/goravel/framework/database/orm"

// SocialAccount links an outside identity ("google:<sub>") to a user.
type SocialAccount struct {
	orm.Model
	Provider string `gorm:"column:provider"`
	Identity string `gorm:"column:identity"`
	Email    string `gorm:"column:email"`
	UserID   uint   `gorm:"column:user_id"`
}

func (SocialAccount) TableName() string { return "social_accounts" }
