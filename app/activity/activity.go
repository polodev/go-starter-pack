// Package activity writes Spatie laravel-activitylog entries (table activity_log), so a Laravel app sharing the
// database reads them too. Every module may log through it; the activity-log module owns the table and the
// admin screens.
//
//	activity.In("users").CausedBy(admin.ID).PerformedOn("user", user.ID).Event("updated").
//		Changes(before, after).LogQuietly("updated user")
package activity

import (
	"encoding/json"
	"strconv"
	"time"

	"goravel/app/facades"
)

// Entry is an activity being built.
type Entry struct {
	logName     string
	causerType  string
	causerID    string
	subjectType string
	subjectID   string
	event       string
	properties  map[string]any
}

// In starts an entry in a log ("users", "roles", "pages" ...).
func In(logName string) *Entry {
	return &Entry{logName: logName, properties: map[string]any{}}
}

// CausedBy is the person who did it (a user id).
func (e *Entry) CausedBy(userID uint) *Entry {
	e.causerType, e.causerID = "user", strconv.FormatUint(uint64(userID), 10)
	return e
}

// PerformedOn is what it was done to.
func (e *Entry) PerformedOn(subjectType string, subjectID uint) *Entry {
	e.subjectType, e.subjectID = subjectType, strconv.FormatUint(uint64(subjectID), 10)
	return e
}

// Event names what happened: created, updated, deleted ...
func (e *Entry) Event(event string) *Entry {
	e.event = event
	return e
}

// With adds a property.
func (e *Entry) With(key string, value any) *Entry {
	e.properties[key] = value
	return e
}

// Changes records what changed, like Spatie: {"attributes": new, "old": old}, only the keys that differ.
// Never pass secrets (passwords, tokens).
func (e *Entry) Changes(before, after map[string]any) *Entry {
	attributes, old := map[string]any{}, map[string]any{}
	for key, value := range after {
		if previous, ok := before[key]; !ok || jsonText(previous) != jsonText(value) {
			attributes[key], old[key] = value, before[key]
		}
	}
	if len(attributes) > 0 {
		e.properties["attributes"], e.properties["old"] = attributes, old
	}
	return e
}

// HasChanges reports whether Changes found anything.
func (e *Entry) HasChanges() bool {
	_, ok := e.properties["attributes"]
	return ok
}

func jsonText(value any) string {
	text, _ := json.Marshal(value)
	return string(text)
}

// Log writes the entry.
func (e *Entry) Log(description string) error {
	properties := jsonText(e.properties)
	now := time.Now().UTC()
	row := map[string]any{
		"log_name": nullable(e.logName), "description": description, "event": nullable(e.event),
		"subject_type": nullable(e.subjectType), "subject_id": nullable(e.subjectID),
		"causer_type": nullable(e.causerType), "causer_id": nullable(e.causerID),
		"properties": properties, "created_at": now, "updated_at": now,
	}
	_, err := facades.DB().Table("activity_log").Insert(row)
	return err
}

// LogQuietly writes the entry; a failure is logged but never breaks the action that was done.
func (e *Entry) LogQuietly(description string) {
	if err := e.Log(description); err != nil {
		facades.Log().Error("activity log: " + err.Error())
	}
}

func nullable(text string) any {
	if text == "" {
		return nil
	}
	return text
}
