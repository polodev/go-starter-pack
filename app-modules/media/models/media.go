// Package models holds the media model.
package models

import (
	"encoding/json"

	"github.com/goravel/framework/database/orm"
)

// Media is one file of the Spatie media table.
type Media struct {
	orm.Model
	ModelType            string  `json:"model_type"`
	ModelID              string  `json:"model_id"`
	UUID                 *string `gorm:"column:uuid" json:"uuid"`
	CollectionName       string  `json:"collection_name"`
	Name                 string  `json:"name"`
	FileName             string  `json:"file_name"`
	MimeType             *string `json:"mime_type"`
	Disk                 string  `json:"disk"`
	ConversionsDisk      *string `json:"conversions_disk"`
	Size                 uint64  `json:"size"`
	Manipulations        string  `json:"-"`
	CustomProperties     string  `json:"-"`
	GeneratedConversions *string `json:"-"`
	ResponsiveImages     string  `json:"-"`
	OrderColumn          *uint   `json:"order_column"`
}

func (Media) TableName() string { return "media" }

// IsPublic reads custom_properties.public; files imported from Laravel without it are private.
func (m Media) IsPublic() bool {
	var properties map[string]any
	_ = json.Unmarshal([]byte(m.CustomProperties), &properties)
	public, _ := properties["public"].(bool)
	return public
}

// Mime is the mime type, or "".
func (m Media) Mime() string {
	if m.MimeType == nil {
		return ""
	}
	return *m.MimeType
}

// KB is the size in kilobytes, rounded up.
func (m Media) KB() uint64 { return (m.Size + 1023) / 1024 }
