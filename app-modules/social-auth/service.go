package socialauth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"

	"goravel/app-modules/auth"
	"goravel/app-modules/social-auth/models"
	"goravel/app/facades"
	appmodels "goravel/app/models"
)

// Provider is the only outside sign-in so far.
const Provider = "google"

// ErrNoAccount: nobody uses this email and GOOGLE_AUTO_CREATE is off.
var ErrNoAccount = errors.New("no account uses this email")

// ErrEmailNotVerified: Google has not verified the address, so it proves nothing.
var ErrEmailNotVerified = errors.New("google has not verified this email")

// FindOrCreateUser turns a Google profile into one of our users, like Laravel Socialite apps usually do:
//
//  1. an account already linked to this Google account;
//  2. else the account with the same email, which is then linked. If that account was never verified, its password
//     is replaced: someone may have registered the address before its owner did, and must not keep a way in;
//  3. else a new, verified account with a password nobody knows (create must be true).
func FindOrCreateUser(profile Profile, create bool) (appmodels.User, error) {
	var user appmodels.User
	if !profile.EmailVerified {
		return user, ErrEmailNotVerified
	}
	identity := Provider + ":" + profile.Subject
	var link models.SocialAccount
	if facades.Orm().Query().Where("identity", identity).First(&link) == nil && link.ID != 0 {
		if facades.Orm().Query().Where("id", link.UserID).First(&user) == nil && user.ID != 0 {
			return user, nil
		}
	}

	email := auth.NormaliseEmail(profile.Email)
	if facades.Orm().Query().Where("email", email).First(&user) == nil && user.ID != 0 {
		if user.EmailVerifiedAt == nil {
			if err := auth.SetPassword(&user, randomPassword()); err != nil {
				return user, err
			}
			if err := auth.MarkVerified(&user); err != nil {
				return user, err
			}
		}
	} else {
		if !create {
			return appmodels.User{}, ErrNoAccount
		}
		name := strings.TrimSpace(profile.Name)
		if name == "" {
			name, _, _ = strings.Cut(email, "@")
		}
		created, taken, err := auth.CreateUser(name, email, randomPassword())
		if taken {
			// Created by a second sign-in at the same moment: use that one.
			if err := facades.Orm().Query().Where("email", email).First(&created); err != nil {
				return created, err
			}
		} else if err != nil {
			return created, err
		}
		user = created
		if user.EmailVerifiedAt == nil {
			if err := auth.MarkVerified(&user); err != nil {
				return user, err
			}
		}
	}

	// A second sign-in at the same moment may have linked it already; the unique identity decides.
	_ = facades.Orm().Query().Create(&models.SocialAccount{Provider: Provider, Identity: identity, Email: email, UserID: user.ID})
	return user, nil
}

// randomPassword is a password nobody knows: the person signs in with Google, an email link or a reset.
func randomPassword() string {
	random := make([]byte, 32)
	_, _ = rand.Read(random)
	return base64.RawURLEncoding.EncodeToString(random)
}
