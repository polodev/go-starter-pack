package pages

import (
	"goravel/app-modules/pages/models"
	"goravel/app/facades"
)

// PagesSeeder adds About, Privacy policy and Terms of service in English and Bangla (skips pages that exist).
type PagesSeeder struct{}

func (s *PagesSeeder) Signature() string { return "PagesSeeder" }

func (s *PagesSeeder) Run() error {
	pages := []struct {
		slug  string
		order int
		texts map[string]Translation
	}{
		{"about", 1, map[string]Translation{
			"en": {Title: "About us", Description: "Who we are and what we do.", Body: "<p>We build reliable web applications.</p><h2>Our promise</h2><ul><li>Clear communication</li><li>Careful work</li></ul>"},
			"bn": {Title: "আমাদের সম্পর্কে", Description: "আমরা কারা এবং কী করি।", Body: "<p>আমরা নির্ভরযোগ্য ওয়েব অ্যাপ্লিকেশন তৈরি করি।</p>"},
		}},
		{"privacy-policy", 2, map[string]Translation{
			"en": {Title: "Privacy policy", Description: "How we handle your data.", Body: "<p>We collect only what we need to run the service, and we never sell your data.</p>"},
			"bn": {Title: "গোপনীয়তা নীতি", Description: "আমরা আপনার তথ্য কীভাবে ব্যবহার করি।", Body: "<p>সেবা চালাতে যা প্রয়োজন শুধু তা-ই আমরা সংগ্রহ করি, এবং আপনার তথ্য কখনো বিক্রি করি না।</p>"},
		}},
		{"terms-of-service", 3, map[string]Translation{
			"en": {Title: "Terms of service", Description: "The rules for using this site.", Body: "<p>Use this site lawfully and respectfully.</p>"},
			"bn": {Title: "সেবার শর্তাবলি", Description: "এই সাইট ব্যবহারের নিয়ম।", Body: "<p>এই সাইট আইনসম্মত ও সম্মানজনকভাবে ব্যবহার করুন।</p>"},
		}},
	}
	for _, item := range pages {
		if _, exists := FindBySlug(item.slug); exists {
			continue
		}
		page := models.Page{Slug: item.slug, IsPublished: true, SortOrder: item.order}
		if err := facades.Orm().Query().Create(&page); err != nil {
			return err
		}
		if err := SaveTranslations(page.ID, item.texts); err != nil {
			return err
		}
	}
	return nil
}
