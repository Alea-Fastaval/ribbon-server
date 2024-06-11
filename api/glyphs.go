package api

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/dreamspawn/ribbon-server/render"
)

func glyphsAPI(sub_path string, vars url.Values, request http.Request) (any, error) {
	glyph_path := resource_dir + "templates/glyphs/"

	if request.Method == "GET" {
		// List all glyphs
		if sub_path == "" {
			data, err := listGlyphs(glyph_path)
			if err != nil {
				// Pass on error message from listGlyphs()
				return data[0], err
			}
			return data, err
		}

		// Load specific glyph
		data := make(map[string]string)
		if var_fg, ok := vars["fg"]; ok {
			data["foreground"] = var_fg[0]
		} else {
			data["foreground"] = "white"
		}

		if var_bg, ok := vars["bg"]; ok {
			data["background"] = var_bg[0]
		} else {
			data["background"] = "black"
		}

		sub_path = strings.ReplaceAll(sub_path, "/", "_") // No directory changing
		glyph_template := render.LoadTemplate("glyphs/" + sub_path)
		output := render.TemplateString(glyph_template, data)

		return map[string]string{
			"content_type": "image/svg+xml",
			"output":       output,
		}, nil
	}

	if request.Method == "POST" {
		file_name := vars["name"][0]
		file_name = strings.ReplaceAll(file_name, "/", "_") // No directory changing
		file_path := glyph_path + file_name + ".tmpl"

		svg := vars["svg"][0]
		svg = strings.ReplaceAll(svg, "var(--glyph-foreground-color)", "{{ .foreground }}")
		svg = strings.ReplaceAll(svg, "var(--glyph-background-color)", "{{ .background }}")

		if _, err := os.Stat(file_path); errors.Is(err, os.ErrNotExist) {
			err := os.WriteFile(file_path, []byte(svg), 0666)
			if err != nil {
				svg = strings.ReplaceAll(svg, "\"", "\\\"")
				return fmt.Sprintf("Could not write glyph file %s with content:\n%s", file_name, svg), err
			}
			return map[string]string{
				"status": "success",
				"file":   file_name,
				"data":   svg,
			}, nil
		}

		return nil, fmt.Errorf("file %s already exist", file_name)
	}

	return nil, fmt.Errorf("endpoint not implemented for method %s", request.Method)
}

func listGlyphs(path string) ([]string, error) {
	file_system := os.DirFS(path)
	matches, err := fs.Glob(file_system, "*.svg.tmpl")

	if err != nil {
		return []string{
			fmt.Sprintf("Could not read content of folder %s\n", path),
		}, err
	}

	return matches, nil
}
