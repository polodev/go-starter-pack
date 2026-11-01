package ses

import "testing"

func TestThePasswordMatchesAWSsDocumentedDerivation(t *testing.T) {
	// Expected values were produced by AWS's own sample script (Python, hmac + sha256).
	cases := map[[2]string]string{
		{"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "us-east-1"}: "BLBM/9hSUELfq8Gw+rU1YcBjkOxGbhT2XG763xVLGWL9",
		{"secret", "ap-southeast-1"}:                              "BAXrMRsGpOLvGtUdNtUIrRXl88mqW9rfYvwNGbVOd49r",
	}
	for input, want := range cases {
		if got := SMTPPassword(input[0], input[1]); got != want {
			t.Errorf("SMTPPassword(%q, %q) = %q, want %q", input[0], input[1], got, want)
		}
	}
	if SMTPHost("eu-west-1") != "email-smtp.eu-west-1.amazonaws.com" {
		t.Error("wrong SES host")
	}
}
