package admin

import (
	"github.com/dreamspawn/ribbon-server/render"
	"github.com/dreamspawn/ribbon-server/server/page"
	"github.com/dreamspawn/ribbon-server/translations"
)

func BuildAdminPage(path string, page *page.Page) {
	data := make(map[string]interface{})
	lang := page.Lang

	translations.Load("admin", lang)
	data["headline"] = translations.Get(lang, "admin", "headline")

	page_tmpl := render.LoadTemplate("admin/main.tmpl")
	content := render.TemplateString(page_tmpl, data)
	page.SetContent(content)
}
