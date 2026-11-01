package controllers

import (
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/http/responses"
	"goravel/app/models"
)

// AuthController is the token API under /api/v1/auth: register, login and the signed-in person.
type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

// The same answer for an unknown email and a wrong password, so the API never reveals who has an account.
const wrongCredentials = "These credentials do not match our records."

// Register creates an account and signs it in: 201 {user, token}.
func (r *AuthController) Register(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]any{
		"name":     "required|max_len:255",
		"email":    "required|email|max_len:255",
		"password": "required|min_len:8|max_len:255",
	})
	if err != nil {
		return responses.ValidationFailed(ctx, map[string][]string{"request": {err.Error()}})
	}
	if validator.Fails() {
		return responses.ValidationFailed(ctx, responses.FromValidator(validator.Errors().All()))
	}

	email := strings.ToLower(strings.TrimSpace(ctx.Request().Input("email")))
	taken, err := facades.Orm().Query().Model(&models.User{}).Where("email", email).Exists()
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": "Server Error"})
	}
	if taken {
		return responses.ValidationFailed(ctx, map[string][]string{"email": {"The email has already been taken."}})
	}

	hash, err := facades.Hash().Make(ctx.Request().Input("password"))
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": "Server Error"})
	}
	user := models.User{Name: strings.TrimSpace(ctx.Request().Input("name")), Email: email, Password: hash}
	if err := facades.Orm().Query().Create(&user); err != nil {
		// Two sign-ups with one address at the same moment: the unique index decides.
		return responses.ValidationFailed(ctx, map[string][]string{"email": {"The email has already been taken."}})
	}
	token, err := facades.Auth(ctx).Login(&user)
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": "Server Error"})
	}
	return ctx.Response().Json(http.StatusCreated, http.Json{"user": user, "token": token})
}

// Login checks the email and password: 200 {user, token}, or 422 with the same message for every failure.
func (r *AuthController) Login(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]any{
		"email":    "required|email",
		"password": "required",
	})
	if err != nil || validator.Fails() {
		if err == nil {
			return responses.ValidationFailed(ctx, responses.FromValidator(validator.Errors().All()))
		}
		return responses.ValidationFailed(ctx, map[string][]string{"request": {err.Error()}})
	}

	var user models.User
	email := strings.ToLower(strings.TrimSpace(ctx.Request().Input("email")))
	found := facades.Orm().Query().Where("email", email).First(&user) == nil && user.ID != 0
	if !found || !facades.Hash().Check(ctx.Request().Input("password"), user.Password) {
		return responses.ValidationFailed(ctx, map[string][]string{"email": {wrongCredentials}})
	}
	token, err := facades.Auth(ctx).Login(&user)
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": "Server Error"})
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"user": user, "token": token})
}

// Me is the signed-in person: 200 {data: user}. The route is behind middleware.Authenticate.
func (r *AuthController) Me(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Auth(ctx).User(&user); err != nil || user.ID == 0 {
		return responses.Unauthenticated(ctx)
	}
	return ctx.Response().Json(http.StatusOK, http.Json{"data": user})
}
