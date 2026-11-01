package pages

import (
	"errors"
	"html/template"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"

	"goravel/app-modules/pages/models"
	"goravel/app/facades"
	"goravel/app/web"
)

// Content is a page in one language, ready to show: each part is that language's text, else the English text.
type Content struct {
	Page        models.Page
	Locale      string
	Title       string
	Description string
	Body        template.HTML // cleaned with bluemonday when it was saved
	Fallback    bool          // some part came from English
}

// Translation is the text a staff member typed for one language.
type Translation struct {
	Title       string
	Description string
	Body        string
}

var policy = bluemonday.UGCPolicy()

// CleanHTML keeps safe formatting (headings, lists, links, emphasis) and removes scripts, styles and event
// handlers. Staff HTML is cleaned before it is stored, never on the way out only.
func CleanHTML(html string) string { return strings.TrimSpace(policy.Sanitize(html)) }

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ValidSlug is lower-case words joined by dashes.
func ValidSlug(slug string) bool { return slugPattern.MatchString(slug) && len(slug) <= 120 }

func translations(pageID uint) map[string]models.PageTranslation {
	var rows []models.PageTranslation
	_ = facades.Orm().Query().Where("page_id", pageID).Find(&rows)
	byLocale := map[string]models.PageTranslation{}
	for _, row := range rows {
		byLocale[row.Locale] = row
	}
	return byLocale
}

// Load is a page's content in a language, each part falling back to English on its own.
func Load(page models.Page, locale string) Content {
	rows := translations(page.ID)
	english, local := rows["en"], rows[locale]
	content := Content{Page: page, Locale: locale}
	pick := func(localValue, englishValue *string) string {
		if text := models.Text(localValue); strings.TrimSpace(text) != "" {
			return text
		}
		if locale != "en" {
			content.Fallback = true
		}
		return models.Text(englishValue)
	}
	content.Title = pick(local.Title, english.Title)
	content.Description = pick(local.Description, english.Description)
	content.Body = template.HTML(pick(local.Body, english.Body)) // stored already cleaned
	return content
}

// FindBySlug finds a page.
func FindBySlug(slug string) (models.Page, bool) {
	var page models.Page
	if facades.Orm().Query().Where("slug", slug).First(&page) != nil || page.ID == 0 {
		return page, false
	}
	return page, true
}

// SaveTranslations stores the texts: English needs a title and a body; another language with every part empty
// is deleted; an empty part is stored as NULL (it falls back to English).
func SaveTranslations(pageID uint, texts map[string]Translation) error {
	english := texts["en"]
	if strings.TrimSpace(english.Title) == "" || strings.TrimSpace(CleanHTML(english.Body)) == "" {
		return errors.New("english needs a title and text")
	}
	for _, locale := range web.Locales {
		text := texts[locale]
		title, description, body := nullable(text.Title), nullable(text.Description), nullable(CleanHTML(text.Body))
		existing := translations(pageID)[locale]
		if locale != "en" && title == nil && description == nil && body == nil {
			if existing.ID != 0 {
				if _, err := facades.Orm().Query().Delete(&existing); err != nil {
					return err
				}
			}
			continue
		}
		row := existing
		row.PageID, row.Locale, row.Title, row.Description, row.Body = pageID, locale, title, description, body
		if err := facades.Orm().Query().Save(&row); err != nil {
			return err
		}
	}
	return nil
}

func nullable(text string) *string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	return &text
}
