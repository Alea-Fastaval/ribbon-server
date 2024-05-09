package server

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/dreamspawn/ribbon-server/admin"
	"github.com/dreamspawn/ribbon-server/config"
	"github.com/dreamspawn/ribbon-server/render"
	"github.com/dreamspawn/ribbon-server/server/page"
	"github.com/dreamspawn/ribbon-server/server/svg"
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
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")

	var query url.Values
	switch request.Method {
	case "GET":
		query = request.URL.Query()
	}

	// io.WriteString(writer, "Fastaval Ribbon Machine\n")
	// path := request.URL.Path
	// query := request.URL.RawQuery

	page := new(page.Page)
	page.Lang = fallback_lang
	page.AddCSS("main.css")

	// Parse template files
	root_tmpl := render.LoadTemplate("root.tmpl")

	admin_page, found := strings.CutPrefix(request.URL.Path, "/"+admin_slug)
	if found && (admin_page == "" || strings.HasPrefix(admin_page, "/")) {
		// Admin pages
		admin.BuildAdminPage(admin_page, page)
		page.AddTitle("[Admin] Fastaval Ribbon Server")
	} else {
		// User pages
		page.SetContent(svg.GetSVGTest(query))
		page.AddTitle("Fastaval Ribbon Server")
	}

	render.WriteTemplate(root_tmpl, writer, page)

	// fmt.Fprintf(writer, "Path: %+v\n", path)
	// fmt.Fprintf(writer, "Query: %+v\n", query)
	// fmt.Fprintf(writer, "Header: %+v\n", request.Header)
	// fmt.Fprintf(writer, "RequestURI: %+v\n", request.RequestURI)
}
