package auth

import (
	"strings"
	"sync"
	"time"

	"github.com/goravel/framework/support/carbon"

	"goravel/app/facades"
	"goravel/app/models"
)

// CreateUser stores a new account with a bcrypt password. taken is true when the email is already used (also
// when two sign-ups race: the unique index decides).
func CreateUser(name, email, password string) (user models.User, taken bool, err error) {
	email = NormaliseEmail(email)
	exists, err := facades.Orm().Query().Model(&models.User{}).Where("email", email).Exists()
	if err != nil {
		return user, false, err
	}
	if exists {
		return user, true, nil
	}
	hash, err := facades.Hash().Make(password)
	if err != nil {
		return user, false, err
	}
	user = models.User{Name: strings.TrimSpace(name), Email: email, Password: hash}
	if err := facades.Orm().Query().Create(&user); err != nil {
		return models.User{}, true, nil
	}
	return user, false, nil
}

var (
	dummyOnce sync.Once
	dummy     string
)

// dummyHash is a real bcrypt hash, so checking an unknown email costs as much as checking a known one.
func dummyHash() string {
	dummyOnce.Do(func() { dummy, _ = facades.Hash().Make("not-a-real-password") })
	return dummy
}

// CheckCredentials finds the account and checks the password. Unknown email and wrong password both answer
// false, and both run a hash check so they take about as long.
func CheckCredentials(email, password string) (models.User, bool) {
	var user models.User
	found := facades.Orm().Query().Where("email", NormaliseEmail(email)).First(&user) == nil && user.ID != 0
	if !found {
		_ = facades.Hash().Check(password, dummyHash())
		return models.User{}, false
	}
	return user, facades.Hash().Check(password, user.Password)
}

// MarkVerified sets email_verified_at.
func MarkVerified(user *models.User) error {
	user.EmailVerifiedAt = carbon.NewDateTime(carbon.FromStdTime(time.Now()))
	return facades.Orm().Query().Save(user)
}

// SetPassword stores a new bcrypt password.
func SetPassword(user *models.User, password string) error {
	hash, err := facades.Hash().Make(password)
	if err != nil {
		return err
	}
	user.Password = hash
	return facades.Orm().Query().Save(user)
}
