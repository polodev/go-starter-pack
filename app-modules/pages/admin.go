package pages

import (
	"strconv"
	"strings"

	"github.com/goravel/framework/contracts/http"

	adminlayout "goravel/app-modules/admin-layout"
	"goravel/app-modules/auth"
	"goravel/app-modules/pages/models"
	"goravel/app/activity"
	"goravel/app/facades"
	"goravel/app/web"
)

var pagesCrumb = adminlayout.Crumb{Label: "Pages", Href: "/admin/pages"}

// AdminIndex is GET /admin/pages.
func AdminIndex(ctx http.Context) http.Response {
	var list []models.Page
	_ = facades.Orm().Query().OrderBy("sort_order").OrderBy("id").Find(&list)
	titles := map[uint]string{}
	languages := map[uint]string{}
	for _, page := range list {
		titles[page.ID] = Load(page, "en").Title
		names := []string{}
		for locale := range translations(page.ID) {
			names = append(names, locale)
		}
		languages[page.ID] = strings.Join(sortedLocales(names), ", ")
	}
	return adminlayout.View(ctx, "pages::admin-index", "Pages", []adminlayout.Crumb{{Label: "Pages"}},
		map[string]any{"pages": list, "titles": titles, "languages": languages})
}

func sortedLocales(names []string) []string {
	var ordered []string
	for _, locale := range web.Locales {
		for _, name := range names {
			if name == locale {
				ordered = append(ordered, locale)
			}
		}
	}
	return ordered
}

func findPage(ctx http.Context) (models.Page, bool) {
	var page models.Page
	id := adminlayout.IDParam(ctx)
	if id == 0 || facades.Orm().Query().Find(&page, id) != nil || page.ID == 0 {
		return page, false
	}
	return page, true
}

// AdminShow is GET /admin/pages/{id}: read-only, every language side by side.
func AdminShow(ctx http.Context) http.Response {
	page, ok := findPage(ctx)
	if !ok {
		return adminlayout.NotFound(ctx)
	}
	title := Load(page, "en").Title
	return adminlayout.View(ctx, "pages::admin-show", title, []adminlayout.Crumb{pagesCrumb, {Label: title}},
		map[string]any{"page": page, "texts": formTexts(page.ID)})
}

// formTexts are the stored texts per language for the form.
func formTexts(pageID uint) map[string]Translation {
	texts := map[string]Translation{}
	for locale, row := range translations(pageID) {
		texts[locale] = Translation{Title: models.Text(row.Title), Description: models.Text(row.Description), Body: models.Text(row.Body)}
	}
	return texts
}

func formValues(page models.Page, texts map[string]Translation, errors map[string]string) map[string]any {
	return map[string]any{"page": page, "texts": texts, "errors": errors, "locales": web.Locales}
}

// AdminCreate is GET /admin/pages/create.
func AdminCreate(ctx http.Context) http.Response {
	return adminlayout.View(ctx, "pages::admin-edit", "New page", []adminlayout.Crumb{pagesCrumb, {Label: "New"}},
		formValues(models.Page{}, map[string]Translation{}, nil))
}

// AdminEdit is GET /admin/pages/{id}/edit.
func AdminEdit(ctx http.Context) http.Response {
	page, ok := findPage(ctx)
	if !ok {
		return adminlayout.NotFound(ctx)
	}
	title := Load(page, "en").Title
	return adminlayout.View(ctx, "pages::admin-edit", "Edit "+title,
		[]adminlayout.Crumb{pagesCrumb, {Label: title, Href: "/admin/pages/" + strconv.Itoa(int(page.ID))}, {Label: "Edit"}},
		formValues(page, formTexts(page.ID), nil))
}

// AdminStore is POST /admin/pages (create) and AdminUpdate is POST /admin/pages/{id}.
func AdminStore(ctx http.Context) http.Response { return save(ctx, models.Page{}) }

func AdminUpdate(ctx http.Context) http.Response {
	page, ok := findPage(ctx)
	if !ok {
		return adminlayout.NotFound(ctx)
	}
	return save(ctx, page)
}

func save(ctx http.Context, page models.Page) http.Response {
	isNew := page.ID == 0
	before := map[string]any{"slug": page.Slug, "is_published": page.IsPublished, "sort_order": page.SortOrder}
	page.Slug = strings.TrimSpace(ctx.Request().Input("slug"))
	page.IsPublished = ctx.Request().Input("is_published") == "1"
	page.SortOrder, _ = strconv.Atoi(ctx.Request().Input("sort_order", "0"))
	texts := map[string]Translation{}
	for _, locale := range web.Locales {
		texts[locale] = Translation{
			Title:       ctx.Request().Input("title_" + locale),
			Description: ctx.Request().Input("description_" + locale),
			Body:        ctx.Request().Input("body_" + locale),
		}
	}
	errors := map[string]string{}
	if !ValidSlug(page.Slug) {
		errors["slug"] = "Use lower-case letters, digits and dashes (about-us)."
	} else {
		taken, _ := facades.Orm().Query().Model(&models.Page{}).Where("slug", page.Slug).Where("id <> ?", page.ID).Exists()
		if taken {
			errors["slug"] = "Another page already uses this address."
		}
	}
	if strings.TrimSpace(texts["en"].Title) == "" || CleanHTML(texts["en"].Body) == "" {
		errors["en"] = "English needs a title and text; other languages are optional."
	}
	if len(errors) > 0 {
		return adminlayout.ViewStatus(ctx, http.StatusUnprocessableEntity, "pages::admin-edit", "Edit page",
			[]adminlayout.Crumb{pagesCrumb, {Label: "Edit"}}, formValues(page, texts, errors))
	}
	if err := facades.Orm().Query().Save(&page); err != nil {
		return ctx.Response().Redirect(http.StatusSeeOther, "/admin/pages?error=save")
	}
	if err := SaveTranslations(page.ID, texts); err != nil {
		return ctx.Response().Redirect(http.StatusSeeOther, "/admin/pages/"+strconv.Itoa(int(page.ID))+"/edit?error=save")
	}
	if user := auth.CurrentUser(ctx); user != nil {
		event, description := "updated", "updated page"
		if isNew {
			event, description = "created", "created page"
		}
		activity.In("pages").CausedBy(user.ID).PerformedOn("page", page.ID).Event(event).
			Changes(before, map[string]any{"slug": page.Slug, "is_published": page.IsPublished, "sort_order": page.SortOrder}).
			LogQuietly(description)
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/admin/pages/"+strconv.Itoa(int(page.ID))+"?saved=page")
}
