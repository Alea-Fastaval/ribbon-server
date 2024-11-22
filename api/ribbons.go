package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dreamspawn/ribbon-server/database"
)

// /api/ribbons
func ribbonsAPI(sub_path string, vars url.Values, request http.Request) (any, error) {
	//--------------------------------------------------------------------------------------------------------------------
	// GET
	//--------------------------------------------------------------------------------------------------------------------
	if request.Method == "GET" {
		ribbons, err := database.GetRibbons()
		if err != nil {
			api_error("could not load ribbons from database", nil)
		}

		result := make(map[uint][]database.Ribbon)
		for _, ribbon := range ribbons {
			result[ribbon.Category] = append(result[ribbon.Category], ribbon)
		}

		return result, nil
	}

	//--------------------------------------------------------------------------------------------------------------------
	// POST
	//--------------------------------------------------------------------------------------------------------------------
	if request.Method == "POST" {
		true_string := map[string]bool{
			"t":    true,
			"true": true,
			"y":    true,
			"yes":  true,
		}

		var cat_id uint64
		if cat, ok := vars["category"]; ok {
			cat_id, _ = strconv.ParseUint(cat[0], 10, 32)
		} else {
			api_error("missing parameter: category", nil)
		}

		var glyph_id uint64
		if glyph, ok := vars["glyph"]; ok {
			glyph_id, _ = strconv.ParseUint(glyph[0], 10, 32)
		} else {
			api_error("missing parameter: glyph", nil)
		}

		var no_wings bool
		if nowings, ok := vars["no_wings"]; ok {
			no_wings = true_string[strings.ToLower(nowings[0])]
		} else {
			no_wings = false
		}

		new_ribbon, err := database.CreateRibbon(
			uint(cat_id),
			uint(glyph_id),
			no_wings,
		)

		if err != nil {
			api_error(fmt.Sprintf("Failed to create new ribbon with values %+v\n", vars), err)
		}

		for key, value := range vars {
			if lang, found := strings.CutPrefix(key, "name_"); found {
				database.AddTranslation(lang, "ribbons."+fmt.Sprint(new_ribbon.ID)+".name", value[0])
			}
			if lang, found := strings.CutPrefix(key, "desc_"); found {
				database.AddTranslation(lang, "ribbons."+fmt.Sprint(new_ribbon.ID)+".desc", value[0])
			}
		}

		return new_ribbon, err
	}

	return nil, fmt.Errorf("endpoint not implemented for method %s", request.Method)
}
