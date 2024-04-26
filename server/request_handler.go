package server

import (
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/dreamspawn/ribbon-server/config"
)

type header map[string]string

type RequestHandler struct {
}

func (handler RequestHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	io.WriteString(writer, "Fastaval Ribbon Machine\n")
	// path := request.URL.Path
	// query := request.URL.RawQuery

	tmpl_dir := config.Get("resource_dir") + "templates/"
	root_tmpl_path := tmpl_dir + "root.tmpl"

	root, err := template.ParseFiles(root_tmpl_path)
	if err != nil {
		fmt.Printf("Could not parse template file %s\n", root_tmpl_path)
		panic(err)
	}

	var headers []header

	title_header := header{
		"Type":  "title",
		"Value": "Fastaval Ribbon Server",
	}
	headers = append(headers, title_header)

	content := "Content Data"
	data := map[string]interface{}{
		"Headers": headers,
		"Content": content,
	}
	err = root.Execute(writer, data)
	if err != nil {
		panic(err)
	}

	// fmt.Fprintf(writer, "Path: %+v\n", path)
	// fmt.Fprintf(writer, "Query: %+v\n", query)
	// fmt.Fprintf(writer, "Header: %+v\n", request.Header)
	// fmt.Fprintf(writer, "RequestURI: %+v\n", request.RequestURI)
}
