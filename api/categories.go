package api

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/dreamspawn/ribbon-server/database"
)

func categoriesAPI(sub_path string, vars url.Values, request http.Request) (any, error) {
	if request.Method == "GET" {
		categories, err := database.GetCategories()
		if err != nil {
			return categories, err
		}
		slices.SortFunc(categories, func(a, b database.Category) int {
			return int(a.Ordering) - int(b.Ordering)
		})

		return categories, nil
	}

	if request.Method == "POST" {
		new_category, err := database.CreateCategory(
			vars["background_color"][0],
			vars["stripes_color"][0],
			vars["glyph_color"][0],
			vars["wing1_color"][0],
			vars["wing2_color"][0],
		)

		if err != nil {
			fmt.Printf("Failed to create new category with values %+v\n", vars)
			panic(err)
		}

		for key, value := range vars {
			lang, found := strings.CutPrefix(key, "name_")
			if found {
				database.AddTranslation(lang, "categories."+fmt.Sprint(new_category.ID), value[0])
			}
		}

		return new_category, err
	}

	return nil, fmt.Errorf("endpoint not implemented for method %s", request.Method)
}
