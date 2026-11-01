// Package responses holds the Laravel-shaped JSON answers of the API, so a Laravel-minded client works unchanged.
package responses

import (
	"fmt"
	"sort"

	"github.com/goravel/framework/contracts/http"
)

// UnauthenticatedBody is Laravel's 401 body.
var UnauthenticatedBody = http.Json{"message": "Unauthenticated."}

// Unauthenticated answers 401 with Laravel's body.
func Unauthenticated(ctx http.Context) http.Response {
	return ctx.Response().Json(http.StatusUnauthorized, UnauthenticatedBody)
}

// ValidationFailed is Laravel's 422 body: the first error (plus "(and N more errors)") and every error by field.
func ValidationFailed(ctx http.Context, errors map[string][]string) http.Response {
	fields := make([]string, 0, len(errors))
	total := 0
	for field, messages := range errors {
		fields = append(fields, field)
		total += len(messages)
	}
	sort.Strings(fields)
	message := "The given data was invalid."
	if len(fields) > 0 {
		message = errors[fields[0]][0]
		if total > 1 {
			message += fmt.Sprintf(" (and %d more error%s)", total-1, plural(total-1))
		}
	}
	return ctx.Response().Json(http.StatusUnprocessableEntity, http.Json{"message": message, "errors": errors})
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// FromValidator turns Goravel's {field: {rule: message}} errors into Laravel's {field: [messages]}.
func FromValidator(all map[string]map[string]string) map[string][]string {
	errors := map[string][]string{}
	for field, rules := range all {
		names := make([]string, 0, len(rules))
		for rule := range rules {
			names = append(names, rule)
		}
		sort.Strings(names)
		for _, rule := range names {
			errors[field] = append(errors[field], rules[rule])
		}
	}
	return errors
}
