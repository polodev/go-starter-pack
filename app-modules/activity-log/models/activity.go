// Package models holds the activity log model.
package models

import (
	"encoding/json"

	"github.com/goravel/framework/database/orm"
)

// Activity is one activity_log row.
type Activity struct {
	orm.Model
	LogName     *string `json:"log_name"`
	Description string  `json:"description"`
	SubjectType *string `json:"subject_type"`
	Event       *string `json:"event"`
	SubjectID   *string `json:"subject_id"`
	CauserType  *string `json:"causer_type"`
	CauserID    *string `json:"causer_id"`
	Properties  *string `json:"properties"`
	BatchUUID   *string `gorm:"column:batch_uuid" json:"batch_uuid"`
}

func (Activity) TableName() string { return "activity_log" }

// Change is one changed field: its old and new value.
type Change struct {
	Field string
	Old   string
	New   string
}

// Changes reads Spatie's {"attributes": {...}, "old": {...}} from the properties.
func (a Activity) Changes() []Change {
	if a.Properties == nil {
		return nil
	}
	var properties struct {
		Attributes map[string]any `json:"attributes"`
		Old        map[string]any `json:"old"`
	}
	if json.Unmarshal([]byte(*a.Properties), &properties) != nil {
		return nil
	}
	var changes []Change
	for field, value := range properties.Attributes {
		changes = append(changes, Change{Field: field, Old: text(properties.Old[field]), New: text(value)})
	}
	return changes
}

// PrettyProperties is the properties JSON, indented.
func (a Activity) PrettyProperties() string {
	if a.Properties == nil {
		return ""
	}
	var value any
	if json.Unmarshal([]byte(*a.Properties), &value) != nil {
		return *a.Properties
	}
	pretty, _ := json.MarshalIndent(value, "", "  ")
	return string(pretty)
}

func text(value any) string {
	if value == nil {
		return "—"
	}
	if s, ok := value.(string); ok {
		return s
	}
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

// Value is a nullable column for templates ("" when NULL).
func Value(field *string) string {
	if field == nil {
		return ""
	}
	return *field
}
