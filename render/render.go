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

var tmpl_dir string

func ConfigReady() {
	tmpl_dir = config.Get("resource_dir") + "templates/"
}

func LoadTemplate(path string) *template.Template {
	template, err := template.ParseFiles(tmpl_dir + path)
	if err != nil {
		fmt.Printf("Could not parse template file %s\n", tmpl_dir+path)
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

func FindSeniorityTemplate(years int) *template.Template {
	seniority_folder := "seniority/"
	path := tmpl_dir + seniority_folder
	files, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("Could not read content of folder %s\n", path)
		panic(err)
	}

	max := 0
	for i := 0; i < len(files); i++ {
		name, found := strings.CutSuffix(files[i].Name(), ".tmpl")
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
		fmt.Printf("Could not find any template for %d years in folder %s\n", years, path)
		return nil
	}

	return LoadTemplate(fmt.Sprintf("%s%d.tmpl", seniority_folder, max))
}
