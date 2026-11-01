// Package models holds the page models.
package models

import "github.com/goravel/framework/database/orm"

// Page is a website page; its text lives in PageTranslation rows.
type Page struct {
	orm.Model
	Slug        string `json:"slug"`
	IsPublished bool   `json:"is_published"`
	SortOrder   int    `json:"sort_order"`
}

// PageTranslation is the text of a page in one language. A NULL part falls back to English.
type PageTranslation struct {
	orm.Model
	PageID      uint    `json:"page_id"`
	Locale      string  `json:"locale"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Body        *string `json:"body"`
}

// Text of a nullable part.
func Text(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
