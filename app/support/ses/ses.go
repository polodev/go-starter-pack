// Package ses holds the one thing Amazon SES needs beyond plain SMTP: its SMTP password, derived from an IAM
// secret access key (AWS's documented algorithm). config/mail.go uses it when MAIL_MAILER=ses, so Goravel's own
// mail (facades.Mail(), make:mail, Queue()) sends through SES with the IAM user's normal keys.
package ses

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func sign(key []byte, message string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(message))
	return mac.Sum(nil)
}

// SMTPPassword derives the SES SMTP password; the user name is the access key id.
func SMTPPassword(secretAccessKey, region string) string {
	signature := sign([]byte("AWS4"+secretAccessKey), "11111111")
	for _, part := range []string{region, "ses", "aws4_request", "SendRawEmail"} {
		signature = sign(signature, part)
	}
	return base64.StdEncoding.EncodeToString(append([]byte{0x04}, signature...))
}

// SMTPHost is the SES SMTP endpoint of a region.
func SMTPHost(region string) string { return "email-smtp." + region + ".amazonaws.com" }
