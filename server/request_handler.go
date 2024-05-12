package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/dreamspawn/ribbon-server/admin"
	"github.com/dreamspawn/ribbon-server/api"
	"github.com/dreamspawn/ribbon-server/config"
	"github.com/dreamspawn/ribbon-server/render"
	"github.com/dreamspawn/ribbon-server/server/page"
	"github.com/dreamspawn/ribbon-server/server/svg"
	"github.com/dreamspawn/ribbon-server/translations"
)

var admin_slug string
var fallback_lang string

func ConfigReady() {
	admin_slug = config.Get("admin_slug")
	fallback_lang = config.Get("fallback_lang")

	page.ConfigReady()
}

type RequestHandler struct {
}

func (handler RequestHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var vars url.Values
	switch request.Method {
	case "GET":
		vars = request.URL.Query()
	case "POST":
		err := request.ParseForm()
		if err != nil {
			fmt.Print("Could not parse form data\n")
			panic(err)
		}
		vars = request.Form
	}

	// io.WriteString(writer, "Fastaval Ribbon Machine\n")
	// path := request.URL.Path
	// query := request.URL.RawQuery

	// Handle API calls
	api_endpoint, found := strings.CutPrefix(request.URL.Path, "/api/")
	if found {
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		json := api.Handle(api_endpoint, vars, request.Method)
		io.WriteString(writer, json)
		return
	}

	// Handle standard page
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")

	page := new(page.Page)
	page.Lang = fallback_lang

	page.AddCSS("theme.css")
	page.AddCSS("main.css")
	page.AddCSS("fontawesome.css")

	page.AddJS("jquery-3.7.1.js")
	page.AddJS("render.js")

	translations.Load("general", page.Lang)

	// Parse template files
	root_tmpl := render.LoadTemplate("root.tmpl")

	// Check for admin slug and make sure it's either seperated with "/" or end of path
	admin_page, found := strings.CutPrefix(request.URL.Path, "/"+admin_slug)
	if found && (admin_page == "" || strings.HasPrefix(admin_page, "/")) {
		// Admin pages
		admin.BuildAdminPage(admin_page, page)
		page.AddTitle("[Admin] Fastaval Ribbon Server")
	} else {
		// User pages
		var page_content string
		headline := translations.Get(page.Lang, "general", "headline")

		main_tmpl := render.LoadTemplate("main-content.tmpl")
		page_content = render.TemplateString(
			main_tmpl,
			map[string]string{
				"headline": headline,
			},
		)

		page_content += svg.GetSVGTest(vars)

		page.SetContent(page_content)
		page.AddTitle("Fastaval Ribbon Server")
	}

	render.WriteTemplate(root_tmpl, writer, page)

	// fmt.Fprintf(writer, "Path: %+v\n", path)
	// fmt.Fprintf(writer, "Query: %+v\n", query)
	// fmt.Fprintf(writer, "Header: %+v\n", request.Header)
	// fmt.Fprintf(writer, "RequestURI: %+v\n", request.RequestURI)
}
