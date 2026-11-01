package feature

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	contractstesting "github.com/goravel/framework/contracts/testing/http"
	"github.com/stretchr/testify/require"

	"goravel/tests"
)

// browser drives the web pages like a person: it keeps cookies (the session), reads the CSRF token from each
// page it opens and sends it with every form.
type browser struct {
	t       *testing.T
	test    *tests.TestCase
	cookies map[string]*http.Cookie
	token   string
}

func newBrowser(t *testing.T, test *tests.TestCase) *browser {
	return &browser{t: t, test: test, cookies: map[string]*http.Cookie{}}
}

var csrfInput = regexp.MustCompile(`name="_token" value="([^"]+)"`)

func (b *browser) remember(response contractstesting.Response) string {
	for _, cookie := range response.Cookies() {
		if cookie.MaxAge < 0 || cookie.Value == "" {
			delete(b.cookies, cookie.Name)
			continue
		}
		b.cookies[cookie.Name] = &http.Cookie{Name: cookie.Name, Value: cookie.Value}
	}
	body, _ := response.Content()
	if match := csrfInput.FindStringSubmatch(body); match != nil {
		b.token = match[1]
	}
	return body
}

func (b *browser) jar() []*http.Cookie {
	jar := make([]*http.Cookie, 0, len(b.cookies))
	for _, cookie := range b.cookies {
		jar = append(jar, cookie)
	}
	return jar
}

// get opens a page; it returns the response and its HTML.
func (b *browser) get(uri string) (contractstesting.Response, string) {
	response, err := b.test.Http(b.t).WithCookies(b.jar()).Get(uri)
	require.NoError(b.t, err)
	return response, b.remember(response)
}

// post sends a form with the CSRF token of the last page.
func (b *browser) post(uri string, form url.Values) (contractstesting.Response, string) {
	return b.postWithToken(uri, form, b.token)
}

func (b *browser) postWithToken(uri string, form url.Values, token string) (contractstesting.Response, string) {
	if form == nil {
		form = url.Values{}
	}
	if token != "" {
		form.Set("_token", token)
	}
	response, err := b.test.Http(b.t).WithCookies(b.jar()).
		WithHeader("Content-Type", "application/x-www-form-urlencoded").
		Post(uri, strings.NewReader(form.Encode()))
	require.NoError(b.t, err)
	return response, b.remember(response)
}

// location is where a redirect points.
func location(response contractstesting.Response) string {
	return response.Headers().Get("Location")
}

// upload sends a multipart form with one file and the CSRF token of the last page.
func (b *browser) upload(uri, field, fileName string, content []byte, fields url.Values) (contractstesting.Response, string) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("_token", b.token)
	for key, values := range fields {
		for _, value := range values {
			_ = writer.WriteField(key, value)
		}
	}
	part, _ := writer.CreateFormFile(field, fileName)
	_, _ = part.Write(content)
	_ = writer.Close()
	response, err := b.test.Http(b.t).WithCookies(b.jar()).
		WithHeader("Content-Type", writer.FormDataContentType()).Post(uri, &body)
	require.NoError(b.t, err)
	return response, b.remember(response)
}
