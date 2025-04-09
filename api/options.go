package api

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/dreamspawn/ribbon-server/database"
)

func optionsAPI(sub_path string, vars url.Values, request http.Request) (any, error) {
	if request.Method == "GET" {
		options, err := database.GetOptions()
		if err != nil {
			return options, err
		}

		return options, nil
	}

	if request.Method == "POST" {
		err := database.SetOption(
			vars.Get("name"),
			vars.Get("value"),
		)

		if err != nil {
			api_error(fmt.Sprintf("Failed to save option with values %+v\n", vars), err)
		}

		return map[string]string{
			"status":  "success",
			"message": "option saved",
		}, nil
	}

	return nil, fmt.Errorf("endpoint not implemented for method %s", request.Method)
}
