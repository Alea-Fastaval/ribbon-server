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

	years, _ := strconv.ParseUint(vars.Get("seniority"), 0, 0)
	second, _ := strconv.ParseUint(vars.Get("second"), 0, 0)
	leader, _ := strconv.ParseUint(vars.Get("leader"), 0, 0)

	// Load glyph
	glyph, err := database.GetGlyph(ribbon.Glyph)
	if err != nil {
		return fmt.Sprintf("Could not read glyph with id:%d from the database", ribbon.Glyph), err
	}

	return map[string]string{
		"content_type": "image/svg+xml",
		"output":       render.Ribbon(*ribbon, *category, *glyph, years, leader, second),
	}, nil
}
