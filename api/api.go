package api

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/dreamspawn/ribbon-server/database"
)

var endpoints = map[string]func(string, url.Values, string) (any, error){
	"categories": categoriesAPI,
}

func Handle(endpoint string, query url.Values, method string) string {
	base, sub_path, _ := strings.Cut(endpoint, "/")

	var data any
	var err error

	if function, ok := endpoints[base]; ok {
		data, err = function(sub_path, query, method)
	} else {
		data, err = database.Query("SELECT * FROM test", nil)
	}

	if err != nil {
		panic(err)
	}

	json_bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	var pretty bytes.Buffer
	err = json.Indent(&pretty, json_bytes, "", "  ")
	if err != nil {
		panic(err)
	}

	return pretty.String()
}
