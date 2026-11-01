package auth

import (
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/http/responses"
	"goravel/app/models"
)

// The token API under /api/v1/auth (guard "api", JWT bearer tokens). Same answer for an unknown email and a
// wrong password, so the API never reveals who has an account.
const wrongCredentials = "These credentials do not match our records."

// ApiRegister creates an account and signs it in: 201 {user, token}.
func ApiRegister(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(registerRules())
	if err != nil {
		return responses.ValidationFailed(ctx, map[string][]string{"request": {err.Error()}})
	}
	if validator.Fails() {
		return responses.ValidationFailed(ctx, responses.FromValidator(validator.Errors().All()))
	}
	user, taken, err := CreateUser(ctx.Request().Input("name"), ctx.Request().Input("email"), ctx.Request().Input("password"))
	if taken {
		return responses.ValidationFailed(ctx, map[string][]string{"email": {"The email has already been taken."}})
	}
	if err != nil {
		return serverError(ctx)
	}
	token, err := facades.Auth(ctx).Guard("api").Login(&user)
	if err != nil {
		return serverError(ctx)
	}
	return ctx.Response().Json(http.StatusCreated, http.Json{"user": user, "token": token})
}

// ApiLogin checks the email and password: 200 {user, token}, or 422 with the same message for every failure.
func ApiLogin(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]any{"email": "required|email", "password": "required"})
	if err != nil {
		return responses.ValidationFailed(ctx, map[string][]string{"request": {err.Error()}})
	}
	if validator.Fails() {
		return responses.ValidationFailed(ctx, responses.FromValidator(validator.Errors().All()))
	}
	user, ok := CheckCredentials(ctx.Request().Input("email"), ctx.Request().Input("password"))
	if !ok {
		return responses.ValidationFailed(ctx, map[string][]string{"email": {wrongCredentials}})
	}
	token, err := facades.Auth(ctx).Guard("api").Login(&user)
	if err != nil {
		return serverError(ctx)
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"user": user, "token": token})
}

// ApiMe is the signed-in person: 200 {data: user}. The route is behind ApiAuthenticate.
func ApiMe(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Auth(ctx).Guard("api").User(&user); err != nil || user.ID == 0 {
		return responses.Unauthenticated(ctx)
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"data": user})
}

func serverError(ctx http.Context) http.Response {
	return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": "Server Error"})
}

func registerRules() map[string]any {
	return map[string]any{
		"name":     "required|max_len:255",
		"email":    "required|email|max_len:255",
		"password": "required|min_len:8|max_len:255",
	}
}

// NormaliseEmail is how emails are stored and looked up.
func NormaliseEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }
