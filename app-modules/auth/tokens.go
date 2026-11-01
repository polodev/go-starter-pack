package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"goravel/app/facades"
)

// Link purposes and how long each link works.
const (
	PurposeVerifyEmail   = "verify_email"
	PurposeResetPassword = "reset_password"
	PurposeSignIn        = "sign_in"
)

var lifetimes = map[string]time.Duration{
	PurposeVerifyEmail:   24 * time.Hour,
	PurposeResetPassword: 60 * time.Minute,
	PurposeSignIn:        15 * time.Minute,
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// IssueToken makes a one-time secret for the user and stores only its SHA-256.
func IssueToken(userID uint, purpose string) (string, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	secret := base64.RawURLEncoding.EncodeToString(random)
	now := time.Now().UTC()
	_, err := facades.DB().Table("user_security_tokens").Insert(map[string]any{
		"user_id": userID, "purpose": purpose, "token_hash": hashSecret(secret),
		"expires_at": now.Add(lifetimes[purpose]), "created_at": now, "updated_at": now,
	})
	return secret, err
}

// tokenRow is a stored link.
type tokenRow struct {
	ID     uint
	UserID uint
}

// PeekToken returns the user of a valid, unused, unexpired link without using it (to show a form).
func PeekToken(secret, purpose string) (uint, bool) {
	var row tokenRow
	err := facades.DB().Table("user_security_tokens").Select("id", "user_id").
		Where("token_hash", hashSecret(secret)).Where("purpose", purpose).
		WhereNull("used_at").Where("expires_at > ?", time.Now().UTC()).First(&row)
	return row.UserID, err == nil && row.ID != 0
}

// ConsumeToken uses a link once. Two requests with the same link cannot both succeed: only the update that
// marks it used wins.
func ConsumeToken(secret, purpose string) (uint, bool) {
	userID, ok := PeekToken(secret, purpose)
	if !ok {
		return 0, false
	}
	result, err := facades.DB().Table("user_security_tokens").
		Where("token_hash", hashSecret(secret)).Where("purpose", purpose).WhereNull("used_at").
		Update("used_at", time.Now().UTC())
	if err != nil || result.RowsAffected != 1 {
		return 0, false
	}
	return userID, true
}

// PruneTokens deletes used and expired links; returns how many.
func PruneTokens() (int64, error) {
	result, err := facades.DB().Table("user_security_tokens").
		Where("used_at IS NOT NULL OR expires_at < ?", time.Now().UTC()).Delete()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}
