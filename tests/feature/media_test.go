package feature

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app-modules/auth"
	"goravel/app-modules/media"
	"goravel/app-modules/media/models"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/facades"
	usermodels "goravel/app/models"
	"goravel/tests"
)

// MediaTestSuite covers the media library: avatar uploads, validation, private and public access, the admin
// screen, and (with keys) Cloudflare R2.
type MediaTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestMediaTestSuite(t *testing.T) {
	suite.Run(t, new(MediaTestSuite))
}

func (s *MediaTestSuite) SetupTest() {
	s.RefreshDatabase()
	_ = facades.Cache().Flush()
	_, err := rolepermission.Sync()
	s.Require().NoError(err)
	auth.SendMail = func(auth.Outgoing) error { return nil }
}

func (s *MediaTestSuite) person(email, role string) usermodels.User {
	user, _, err := auth.CreateUser("Person", email, "secret-password")
	s.Require().NoError(err)
	s.Require().NoError(auth.MarkVerified(&user))
	if role != "" {
		s.Require().NoError(rolepermission.AssignRole(user.ID, role))
	}
	return user
}

func (s *MediaTestSuite) signedIn(email string) *browser {
	b := newBrowser(s.T(), &s.TestCase)
	b.get("/login")
	response, _ := b.post("/login", url.Values{"email": {email}, "password": {"secret-password"}})
	response.AssertStatus(303)
	return b
}

func pngBytes(shade uint8) []byte {
	picture := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for x := 0; x < 4; x++ {
		for y := 0; y < 4; y++ {
			picture.Set(x, y, color.RGBA{R: shade, G: 100, B: 200, A: 255})
		}
	}
	var buffer bytes.Buffer
	_ = png.Encode(&buffer, picture)
	return buffer.Bytes()
}

func (s *MediaTestSuite) avatars(userID uint) []models.Media {
	return media.Get(media.UserOwner(userID), "avatar")
}

func (s *MediaTestSuite) TestAnAvatarIsStoredServedToAnyoneAndReplacedByTheNextOne() {
	user := s.person("pic@example.com", "")
	b := s.signedIn("pic@example.com")
	b.get("/account/avatar")
	saved, _ := b.upload("/account/avatar", "avatar", "My Photo!.PNG", pngBytes(10), nil)
	s.Equal("/account/avatar?status=saved", location(saved))

	items := s.avatars(user.ID)
	s.Require().Len(items, 1)
	first := items[0]
	s.Equal("my-photo-.png", first.FileName, "a safe file name")
	s.Equal("image/png", first.Mime())
	s.True(first.IsPublic())
	s.Equal("public", first.Disk)
	s.Equal(strconv.Itoa(int(user.ID)), first.ModelID)

	served, err := s.Http(s.T()).Get(media.URL(first))
	s.Require().NoError(err)
	served.AssertStatus(200)
	served.AssertHeader("Content-Type", "image/png")
	body, _ := served.Content()
	s.Equal(string(pngBytes(10)), body, "anyone gets the public file")

	_, profile := b.get("/account/profile")
	s.Contains(profile, `data-testid="profile-avatar"`)

	b.get("/account/avatar")
	b.upload("/account/avatar", "avatar", "second.png", pngBytes(200), nil)
	items = s.avatars(user.ID)
	s.Require().Len(items, 1, "a single collection keeps one file")
	s.NotEqual(first.ID, items[0].ID)
	s.False(facades.Storage().Disk("public").Exists(media.Path(first)), "the old file is deleted")
	gone, _ := s.Http(s.T()).Get(media.URL(first))
	gone.AssertStatus(404)
}

func (s *MediaTestSuite) TestAWrongTypeOrTooLargeFileIsRefusedAndNothingIsStored() {
	user := s.person("strict@example.com", "")
	b := s.signedIn("strict@example.com")
	b.get("/account/avatar")
	wrong, body := b.upload("/account/avatar", "avatar", "notes.txt", []byte("just text"), nil)
	wrong.AssertStatus(422)
	s.Contains(body, `data-testid="upload-error"`)
	big, _ := b.upload("/account/avatar", "avatar", "huge.png", append(pngBytes(1), bytes.Repeat([]byte{0}, 3<<20)...), nil)
	big.AssertStatus(422)
	s.Empty(s.avatars(user.ID), "nothing stored after a refusal")
}

func (s *MediaTestSuite) TestPrivateFilesAreOnlyForStaffWithMediaView() {
	s.person("admin@example.com", "admin")
	s.person("bd@example.com", "business_development")
	s.person("customer@example.com", "")
	admin := s.signedIn("admin@example.com")
	admin.get("/admin/media")
	uploaded, _ := admin.upload("/admin/media", "file", "contract.pdf", []byte("%PDF-1.4 private"), url.Values{"visibility": {"private"}})
	s.Equal("/admin/media?saved=uploaded", location(uploaded))
	var item models.Media
	s.Require().NoError(facades.Orm().Query().Where("collection_name", "files").First(&item))
	s.False(item.IsPublic())
	s.Equal("local", item.Disk, "private files never go to a public disk")
	address := media.URL(item)

	guest, _ := s.Http(s.T()).Get(address)
	guest.AssertStatus(404)
	customer, _ := s.signedIn("customer@example.com").get(address)
	customer.AssertStatus(404)
	staff, body := s.signedIn("bd@example.com").get(address)
	staff.AssertStatus(200)
	s.Equal("%PDF-1.4 private", body)
	staff.AssertHeader("Cache-Control", "private, max-age=0, no-store")
	wrongName, _ := s.signedIn("bd@example.com").get("/media/" + strconv.Itoa(int(item.ID)) + "/other.pdf")
	wrongName.AssertStatus(404)

	bd := s.signedIn("bd@example.com")
	_, list := bd.get("/admin/media")
	s.Contains(list, "contract.pdf")
	s.NotContains(list, `data-testid="delete-media"`, "view only")
	refused, _ := bd.post("/admin/media/"+strconv.Itoa(int(item.ID))+"/delete", nil)
	refused.AssertStatus(403)

	admin.get("/admin/media")
	deleted, _ := admin.post("/admin/media/"+strconv.Itoa(int(item.ID))+"/delete", nil)
	s.Equal("/admin/media?saved=deleted", location(deleted))
	count, _ := facades.Orm().Query().Model(&models.Media{}).Count()
	s.Equal(int64(0), count)
	s.False(facades.Storage().Disk("local").Exists(media.Path(item)))
}

func (s *MediaTestSuite) TestWithR2PublicFilesUseThePublicBucketAndPrivateOnesThePrivateBucket() {
	for _, name := range []string{"R2_ACCESS_KEY_ID", "R2_SECRET_ACCESS_KEY", "R2_ENDPOINT", "R2_PUBLIC_BUCKET", "R2_PRIVATE_BUCKET", "R2_PUBLIC_URL"} {
		if env(name) == "" {
			s.T().Skipf("%s is not set in .env", name)
		}
	}
	s.T().Setenv("MEDIA_STORAGE", "r2")
	s.T().Setenv("MEDIA_PREFIX", env("R2_ROOT")+"/tests/media")
	user := s.person("r2@example.com", "")

	public, err := media.AddBytes(media.UserOwner(user.ID), "avatar", "face.png", "image/png", pngBytes(50))
	s.Require().NoError(err)
	defer func() { _ = media.Delete(public) }()
	s.Equal("r2_public", public.Disk)
	address := media.URL(public)
	s.True(strings.HasPrefix(address, strings.TrimSuffix(env("R2_PUBLIC_URL"), "/")), address)
	response, err := http.Get(address)
	s.Require().NoError(err)
	content, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	s.Equal(http.StatusOK, response.StatusCode, "anyone can read it from R2")
	s.Equal(string(pngBytes(50)), string(content))

	private, err := media.AddBytes(media.Owner{Type: "library", ID: "0"}, "files", "secret.pdf", "application/pdf", []byte("%PDF private"))
	s.Require().NoError(err)
	defer func() { _ = media.Delete(private) }()
	s.Equal("r2_private", private.Disk)
	s.True(strings.HasPrefix(media.URL(private), "/media/"), "private files go through the app")
	anonymous, err := http.Get(strings.TrimSuffix(env("R2_ENDPOINT"), "/") + "/" + env("R2_PRIVATE_BUCKET") + "/" + media.Path(private))
	s.Require().NoError(err)
	_ = anonymous.Body.Close()
	s.NotEqual(http.StatusOK, anonymous.StatusCode, "the private bucket refuses anonymous reads")
	s.person("staff@example.com", "business_development")
	served, body := s.signedIn("staff@example.com").get(media.URL(private))
	served.AssertStatus(200)
	s.Equal("%PDF private", body)
}
