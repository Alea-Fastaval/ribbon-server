package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/dreamspawn/ribbon-server/database"
	"github.com/dreamspawn/ribbon-server/render"
)

func glyphsAPI(sub_path string, vars url.Values, request http.Request) (any, error) {
	if request.Method == "GET" {
		// List all glyphs
		if sub_path == "" {
			data, err := listGlyphs()
			if err != nil {
				// Pass on error message from listGlyphs()
				return data[0], err
			}
			return data, err
		}

		glyph_id, err := strconv.ParseUint(sub_path, 10, 64)
		if err != nil {
			return fmt.Sprintf("Could not convert \"%s\" into an unsigned int for id", sub_path), err
		}

		glyph, err := database.GetGlyph(uint(glyph_id))
		if err != nil {
			return fmt.Sprintf("Could not read glyph with id:%d from the database", glyph_id), err
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

		glyph_template := render.LoadTemplate(glyph.File)
		output := render.TemplateString(glyph_template, data)

		return map[string]string{
			"content_type": "image/svg+xml",
			"output":       output,
		}, nil
	}

	if request.Method == "POST" {
		file_name := vars["name"][0]
		file_name = strings.ReplaceAll(file_name, "/", "_") // No directory changing
		file_path := "glyphs/" + file_name + ".tmpl"
		template_folder := resource_dir + "templates/"

		svg := vars["svg"][0]
		svg = strings.ReplaceAll(svg, "var(--glyph-foreground-color)", "{{ .foreground }}")
		svg = strings.ReplaceAll(svg, "var(--glyph-background-color)", "{{ .background }}")

		if _, err := os.Stat(template_folder + file_path); errors.Is(err, os.ErrNotExist) {
			err := os.WriteFile(template_folder+file_path, []byte(svg), 0666)
			if err != nil {
				svg = strings.ReplaceAll(svg, "\"", "\\\"")
				return fmt.Sprintf("Could not write glyph file %s with content:\n%s", file_name, svg), err
			}

			glyph, err := database.CreateGlyph(file_path)

			if err != nil {
				return "Could not add glyph to database", err
			}

			return map[string]string{
				"status": "success",
				"id":     fmt.Sprintf("%d", glyph.ID),
				"file":   glyph.File,
				"data":   svg,
			}, nil
		}

		return nil, fmt.Errorf("file %s already exist", file_name)
	}

	return nil, fmt.Errorf("endpoint not implemented for method %s", request.Method)
}

func listGlyphs() (map[uint]string, error) {
	glyphs, err := database.GetGlyphs()
	if err != nil {
		return map[uint]string{
			0: "Could not read glyphs from database\n",
		}, err
	}

	result := make(map[uint]string)
	for id, glyph := range glyphs {
		result[id] = glyph.File
	}

	return result, nil
}
