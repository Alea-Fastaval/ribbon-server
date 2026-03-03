package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dreamspawn/ribbon-server/database"
	"github.com/dreamspawn/ribbon-server/server/session"
	"github.com/dreamspawn/ribbon-server/translations"
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
		//Get single ribbon with translations
		if sub_section != "" {
			ribbon_id, err := strconv.ParseUint(sub_section, 10, 32)
			if err != nil {
				api_error(fmt.Sprintf("Could not parse %s as a ribbon ID\n", sub_section), err)
			}

			ribbon, err := database.GetRibbon(uint(ribbon_id))
			if err != nil {
				api_error(fmt.Sprintf("Could not load ribbon with ID:%d\n", ribbon_id), err)
			}

			text := make(map[string]any)
			for _, lang := range translations.GetLanguages() {
				result, err := database.GetTranslation(lang, fmt.Sprintf("ribbons.%d.*", ribbon_id))
				if err != nil {
					api_error(fmt.Sprintf("Could not load translations %s for ribbon %d\n", lang, ribbon_id), err)
				}
				if result_ribbons, ok := result["ribbons"].(map[string]any); ok {
					if result_texts, ok := result_ribbons[strconv.FormatUint(ribbon_id, 10)]; ok {
						text[lang] = result_texts
					}
				}
			}

			return map[string]any{
				"status": "success",
				"ribbon": ribbon,
				"text":   text,
			}, nil
		}

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
		if sub_section == "show" {
			ribbon_id, err := strconv.ParseUint(sub_adress, 10, 32)
			if err != nil {
				api_error(fmt.Sprintf("Could not parse %s as a ribbon ID\n", sub_adress), err)
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

		// Create ribbon
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

		true_string := map[string]bool{
			"t":    true,
			"true": true,
			"y":    true,
			"yes":  true,
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
	// PATCH
	//--------------------------------------------------------------------------------------------------------------------
	if request.Method == "PATCH" {
		var ribbon_id uint64
		if id, ok := vars["id"]; ok {
			ribbon_id, _ = strconv.ParseUint(id[0], 10, 32)
		} else {
			api_error("missing parameter: id", nil)
		}

		if sub_section == "retire" {
			var retired bool
			if value, ok := vars["value"]; ok {
				retired, _ = strconv.ParseBool(value[0])
			} else {
				api_error("missing parameter: value", nil)
			}

			err := database.RetireRibbon(uint(ribbon_id), retired)
			if err != nil {
				api_error(fmt.Sprintf("error trying to retire/unretire ribbon with ID: %d\n", ribbon_id), err)
			}

			return map[string]string{
				"status":  "success",
				"message": fmt.Sprintf("ribbon %d retired status is now %t", ribbon_id, retired),
			}, nil
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

		true_string := map[string]bool{
			"t":    true,
			"true": true,
			"y":    true,
			"yes":  true,
		}

		var no_wings bool
		if nowings, ok := vars["no_wings"]; ok {
			no_wings = true_string[strings.ToLower(nowings[0])]
		} else {
			no_wings = false
		}

		result, err := database.UpdateRibbon(
			uint(ribbon_id),
			uint(cat_id),
			uint(glyph_id),
			no_wings,
		)

		if err != nil {
			api_error(fmt.Sprintf("Failed to update ribbon with values %+v\n", vars), err)
		}

		for key, value := range vars {
			if lang, found := strings.CutPrefix(key, "name_"); found {
				database.UpdateTranslation(lang, "ribbons."+fmt.Sprint(ribbon_id)+".name", value[0])
			}
			if lang, found := strings.CutPrefix(key, "desc_"); found {
				database.UpdateTranslation(lang, "ribbons."+fmt.Sprint(ribbon_id)+".desc", value[0])
			}
		}

		return result, err
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
