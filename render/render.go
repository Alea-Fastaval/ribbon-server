package render

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/dreamspawn/ribbon-server/config"
)

var tmpl_folder string

func ConfigReady() {
	tmpl_folder = config.Get("resource_dir") + "templates/"
}

func LoadTemplate(path string) *template.Template {
	template, err := template.ParseFiles(tmpl_folder + path)
	if err != nil {
		fmt.Printf("Could not parse template file %s\n", tmpl_folder+path)
		panic(err)
	}

	return template
}

func WriteTemplate(template *template.Template, writer io.Writer, data any) {
	err := template.Execute(writer, data)
	if err != nil {
		panic(err)
	}
}

func TemplateString(template *template.Template, data any) string {
	var string_builder strings.Builder

	err := template.Execute(&string_builder, data)
	if err != nil {
		panic(err)
	}

	return string_builder.String()
}

func FindYearTemplate(years int, sub_folder string) *template.Template {
	path := tmpl_folder + sub_folder + "/"
	files, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("Could not read content of folder %s\n", path)
		panic(err)
	}

	max := 0
	for _, file := range files {
		name, found := strings.CutSuffix(file.Name(), ".tmpl")
		if !found {
			continue
		}

		file_years, err := strconv.ParseInt(name, 0, 0)
		if err != nil {
			continue
		}

		file_years_int := int(file_years)
		if file_years_int > max && file_years_int <= years {
			max = file_years_int
		}
	}

	if max == 0 {
		return nil
	}

	return LoadTemplate(fmt.Sprintf("%s/%d.tmpl", sub_folder, max))
}
