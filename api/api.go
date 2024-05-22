package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/dreamspawn/ribbon-server/config"
)

var resource_dir string

func ConfigReady() {
	resource_dir = config.Get("resource_dir")
}

var endpoints = map[string]func(string, url.Values, http.Request) (any, error){
	"categories":   categoriesAPI,
	"translations": translationsAPI,
	"glyphs":       glyphsAPI,
}

func Handle(endpoint string, vars url.Values, request http.Request) string {
	base, sub_path, _ := strings.Cut(endpoint, "/")

	var data any
	var err error

	if function, ok := endpoints[base]; ok {
		data, err = function(sub_path, vars, request)
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
