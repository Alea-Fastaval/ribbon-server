package admin

import (
	"github.com/dreamspawn/ribbon-server/render"
	"github.com/dreamspawn/ribbon-server/server/page"
	"github.com/dreamspawn/ribbon-server/translations"
	"github.com/dreamspawn/ribbon-server/user"
)

func BuildAdminPage(path string, page *page.Page, user user.User) {
	if !user.IsAdmin {
		// Access denied
		translations.Load("general", page.Lang)
		page_tmpl := render.LoadTemplate("admin/no-access.tmpl")
		content := render.TemplateString(page_tmpl, map[string]string{
			"headline": translations.Get(page.Lang, "general", "no_access_headline"),
			"message":  translations.Get(page.Lang, "general", "no_access_message"),
		})
		page.SetContent(content)
		return
	}

	translations.Load("admin", page.Lang)

	page.AddCSS("admin.css")
	page.AddJS("admin.js")

	page_tmpl := render.LoadTemplate("admin/main.tmpl")
	content := render.TemplateString(page_tmpl, map[string]string{
		"headline": translations.Get(page.Lang, "admin", "headline"),
	})

	page.SetContent(content)
}
