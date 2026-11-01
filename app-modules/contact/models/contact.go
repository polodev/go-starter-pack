// Package models holds the contact message model.
package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

// Statuses of a message: new until someone has answered it.
const (
	StatusNew     = "new"
	StatusReplied = "replied"
)

// Statuses lists them for filters and the status form.
var Statuses = []string{StatusNew, StatusReplied}

// Contact is one message from the contact form (contacts).
type Contact struct {
	orm.Model
	Name      string     `gorm:"column:name"`
	Email     string     `gorm:"column:email"`
	Subject   *string    `gorm:"column:subject"`
	Message   string     `gorm:"column:message"`
	Status    string     `gorm:"column:status"`
	RepliedAt *time.Time `gorm:"column:replied_at"`
	Locale    string     `gorm:"column:locale"`
	IPAddress *string    `gorm:"column:ip_address"`
	UserAgent *string    `gorm:"column:user_agent"`
}

func (Contact) TableName() string { return "contacts" }

// SubjectText is the subject, or "" when none was given.
func (c Contact) SubjectText() string {
	if c.Subject == nil {
		return ""
	}
	return *c.Subject
}
