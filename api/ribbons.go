package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dreamspawn/ribbon-server/database"
)

func ribbonsAPI(sub_path string, vars url.Values, request http.Request) (any, error) {
	if request.Method == "POST" {
		true_string := map[string]bool{
			"t":    true,
			"true": true,
			"y":    true,
			"yes":  true,
		}

		cat_id, _ := strconv.ParseUint(vars["category"][0], 10, 32)
		glyph_id, _ := strconv.ParseUint(vars["glyph"][0], 10, 32)

		new_ribbon, err := database.CreateRibbon(
			uint(cat_id),
			uint(glyph_id),
			true_string[strings.ToLower(vars["no_wings"][0])],
		)

		if err != nil {
			fmt.Printf("Failed to create new category with values %+v\n", vars)
			panic(err)
		}

		for key, value := range vars {
			if lang, found := strings.CutPrefix(key, "name_"); found {
				database.AddTranslation(lang, "ribbon."+fmt.Sprint(new_ribbon.ID)+".name", value[0])
			}
			if lang, found := strings.CutPrefix(key, "desc_"); found {
				database.AddTranslation(lang, "ribbon."+fmt.Sprint(new_ribbon.ID)+".desc", value[0])
			}
		}

		return new_ribbon, err
	}

	return nil, fmt.Errorf("endpoint not implemented for method %s", request.Method)
}
