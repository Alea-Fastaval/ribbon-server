package admin

import (
	"github.com/dreamspawn/ribbon-server/render"
	"github.com/dreamspawn/ribbon-server/server/page"
	"github.com/dreamspawn/ribbon-server/translations"
)

func BuildAdminPage(path string, page *page.Page) {
	translations.Load("admin", page.Lang)

	page.AddCSS("admin.css")
	page.AddJS("admin.js")

	page_tmpl := render.LoadTemplate("admin/main.tmpl")
	content := render.TemplateString(page_tmpl, map[string]string{
		"headline": translations.Get(page.Lang, "admin", "headline"),
	})

	page.SetContent(content)
}
