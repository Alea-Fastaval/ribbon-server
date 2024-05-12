package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

var endpoints = map[string]func(string, url.Values, string) (any, error){
	"categories":   categoriesAPI,
	"translations": translationsAPI,
}

func Handle(endpoint string, vars url.Values, method string) string {
	base, sub_path, _ := strings.Cut(endpoint, "/")

	var data any
	var err error

	if function, ok := endpoints[base]; ok {
		data, err = function(sub_path, vars, method)
	} else {
		panic(fmt.Errorf("endpoint %s is not defined", base))
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
