package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dreamspawn/ribbon-server/database"
	"github.com/dreamspawn/ribbon-server/server/session"
)

// /api/ribbons
func ribbonsAPI(sub_path string, vars url.Values, request http.Request) (any, error) {
	sub_section, sub_adress, _ := strings.Cut(sub_path, "/")
	if sub_section == "svg" {
		return svgAPI(sub_adress, vars, request)
	}

	current_session := session.Check(request)
	user := current_session.GetUser()

	//--------------------------------------------------------------------------------------------------------------------
	// GET
	//--------------------------------------------------------------------------------------------------------------------
	if request.Method == "GET" {
		ribbons, err := database.GetRibbons(user.IsAdmin)
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
		//Show ribbon again
		if sub_path[0:5] == "show/" {
			ribbon_id, err := strconv.ParseUint(sub_path[5:], 10, 32)
			if err != nil {
				api_error(fmt.Sprintf("Could not parse %s as a ribbon ID\n", sub_path[5:]), err)
			}

			err = database.ShowRibbon(uint(ribbon_id))
			if err != nil {
				api_error(fmt.Sprintf("error trying to show ribbon with ID: %d\n", ribbon_id), err)
			}

			return map[string]string{
				"status":  "success",
				"message": fmt.Sprintf("ribbon %d is now shown again", ribbon_id),
			}, nil
		}

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

	//--------------------------------------------------------------------------------------------------------------------
	// DELETE
	//--------------------------------------------------------------------------------------------------------------------
	if request.Method == "DELETE" {
		ribbon_id, err := strconv.ParseUint(sub_path, 10, 32)
		if err != nil {
			api_error("missing ribbon id in url", err)
		}

		err = database.HideRibbon(uint(ribbon_id))
		if err != nil {
			api_error(fmt.Sprintf("error trying to hide ribbon with ID: %d\n", ribbon_id), err)
		}

		return map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("ribbon %d is now hidden", ribbon_id),
		}, nil
	}

	return nil, fmt.Errorf("endpoint not implemented for method %s", request.Method)
}
