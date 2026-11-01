package rolepermission

// The roles and permissions are defined here, in code (like config/roles.php and config/permissions.php), and
// copied into the tables by `role-permission:sync`. Edit, then run the sync.

// Permissions by area: key -> description. Keys are "<area>.<action>"; "<area>.*" in a role grants every action
// of the area.
var Permissions = map[string]map[string]string{
	"dashboard": {
		"dashboard.view":   "View the admin dashboard",
		"dashboard.manage": "Manage dashboard settings",
	},
	"users": {
		"users.view":   "View user accounts",
		"users.manage": "Create, edit and block user accounts",
	},
	"roles": {
		"roles.view":   "View role and permission definitions",
		"roles.assign": "Assign roles and direct permissions to users",
		"roles.sync":   "Sync the defined roles and permissions into the database",
	},
	"pages": {
		"pages.view":   "View website pages",
		"pages.manage": "Edit website pages",
	},
	"media": {
		"media.view":   "View uploaded files",
		"media.manage": "Upload and delete files",
	},
	"activity_logs": {
		"activity_logs.view":   "View the activity log",
		"activity_logs.manage": "Delete and clean activity log entries",
	},
	"ai_api_logs": {
		"ai_api_logs.view":   "View AI API logs",
		"ai_api_logs.manage": "Delete AI API logs",
	},
	"queue": {
		"queue.view":   "View jobs, failed jobs and the schedule",
		"queue.manage": "Retry and delete jobs",
	},
}

// RoleDefinition is one staff role. Rank orders the roles; a higher rank includes everything below it.
type RoleDefinition struct {
	Slug        string
	Name        string
	Rank        int
	Description string
	Permissions []string
}

func with(base []string, more ...string) []string {
	return append(append([]string{}, base...), more...)
}

var (
	newEmployee         = []string{"dashboard.view"}
	businessDevelopment = with(newEmployee, "users.view", "pages.view", "media.view")
	employee            = with(businessDevelopment, "pages.manage", "activity_logs.view", "ai_api_logs.view", "queue.view")
	admin               = with(employee, "dashboard.manage", "users.manage", "media.manage", "activity_logs.manage",
		"ai_api_logs.manage", "queue.*")
	developer = with(admin, "roles.*")
)

// Roles in ascending authority, like the Laravel apps: only a developer may assign roles or run the sync.
var Roles = []RoleDefinition{
	{"new_employee", "New Employee", 1, "Admin overview only.", newEmployee},
	{"business_development", "Business Development", 2, "Looks at users, pages and files.", businessDevelopment},
	{"employee", "Employee", 3, "Edits pages and reads the operational logs.", employee},
	{"admin", "Admin", 4, "Manages the platform, but cannot assign roles.", admin},
	{"developer", "Developer", 5, "Everything, including role assignment and the sync.", developer},
}

// AllPermissionKeys lists every defined key.
func AllPermissionKeys() []string {
	var keys []string
	for _, group := range Permissions {
		for key := range group {
			keys = append(keys, key)
		}
	}
	return keys
}
