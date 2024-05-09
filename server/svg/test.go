package svg

import (
	"net/url"
	"strconv"
	"text/template"

	"github.com/dreamspawn/ribbon-server/render"
)

func GetSVGTest(query url.Values) string {
	svg_tmpl := render.LoadTemplate("svg-test.tmpl")
	var seniority_tmpl *template.Template

	years, err := strconv.ParseInt(query.Get("seniority"), 0, 0)
	if err == nil {
		seniority_tmpl = render.FindSeniorityTemplate(int(years))
	} else {
		seniority_tmpl = render.LoadTemplate("seniority/1.tmpl")
	}

	seniority_data := map[string]interface{}{
		"background": "blue",
		"fill_color": "green",
	}
	seniority := render.TemplateString(seniority_tmpl, seniority_data)

	svg_data := map[string]interface{}{
		"background": "blue",
		"seniority":  seniority,
	}

	return render.TemplateString(svg_tmpl, svg_data)
}
