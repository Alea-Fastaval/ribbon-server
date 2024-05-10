package api

import (
	"fmt"
	"net/url"

	"github.com/dreamspawn/ribbon-server/database"
)

func categoriesAPI(sub_path string, vars url.Values, method string) (any, error) {
	if method == "GET" {
		return database.GetCategories()
	}

	return nil, fmt.Errorf("endpoint not implemented for method %s", method)
}
