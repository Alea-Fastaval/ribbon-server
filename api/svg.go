package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dreamspawn/ribbon-server/database"
	"github.com/dreamspawn/ribbon-server/render"
)

// /api/ribbons/svg
func svgAPI(sub_path string, vars url.Values, request http.Request) (any, error) {
	if request.Method != "GET" {
		return nil, fmt.Errorf("endpoint not implemented for method %s", request.Method)
	}

	// Load ribbon
	ribbon_id, err := strconv.ParseUint(sub_path, 10, 32)
	if err != nil {
		api_error("missing ribbon id in url", err)
	}
	ribbon, err := database.GetRibbon(uint(ribbon_id))
	if ribbon == nil {
		api_error(fmt.Sprintf("could not load ribbon with ID: %d", ribbon_id), err)
	}

	// Load category
	category, err := ribbon.GetCategory()
	if err != nil {
		api_error(fmt.Sprintf("could not load category for ribbon with ID: %d", ribbon_id), err)
	}

	svg_data := map[string]interface{}{
		"background": category.Background,
	}

	// Render stripes
	years, err := strconv.ParseInt(vars.Get("seniority"), 0, 0)
	if err == nil {
		seniority_tmpl := render.FindYearTemplate(int(years), "seniority")

		if seniority_tmpl != nil {
			seniority_data := map[string]interface{}{
				"fill_color": category.Stripes,
				"background": category.Background,
			}
			svg_data["seniority"] = render.TemplateString(seniority_tmpl, seniority_data)
		}
	}

	// Render wings
	years = 0
	wing_color := ""
	second, err := strconv.ParseInt(vars.Get("second"), 0, 0)
	if err == nil {
		years = second
		wing_color = category.Wing2
	}
	leader, err := strconv.ParseInt(vars.Get("leader"), 0, 0)
	if err == nil {
		years += leader
		wing_color = category.Wing1
	}

	if years > 0 {
		wing_tmpl := render.FindYearTemplate(int(years), "wings")

		if wing_tmpl != nil {
			wing_data := map[string]interface{}{
				"foreground": wing_color,
			}
			svg_data["wings"] = render.TemplateString(wing_tmpl, wing_data)
		}
	}

	// Load glyph
	glyph, err := database.GetGlyph(ribbon.Glyph)
	if err != nil {
		return fmt.Sprintf("Could not read glyph with id:%d from the database", ribbon.Glyph), err
	}

	// Render glyph
	glyph_data := map[string]string{
		"foreground": category.Glyph,
		"background": category.Background,
	}
	glyph_template := render.LoadTemplate(glyph.File)
	svg_data["glyph"] = render.TemplateString(glyph_template, glyph_data)

	// Render ribbon
	svg_tmpl := render.LoadTemplate("full-ribbon.tmpl")
	return map[string]string{
		"content_type": "image/svg+xml",
		"output":       render.TemplateString(svg_tmpl, svg_data),
	}, nil
}
