package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

// User is a row of Laravel's users table. The password hash is never sent out.
type User struct {
	orm.Model
	Name            string           `json:"name"`
	Email           string           `json:"email"`
	EmailVerifiedAt *carbon.DateTime `json:"email_verified_at"`
	Password        string           `json:"-"`
	RememberToken   *string          `json:"-"`
}
