package api

import (
	"fmt"
	"net/url"
	"os"
)

func glyphsAPI(sub_path string, vars url.Values, method string) (any, error) {
	glyph_path := resource_dir + "public/glyphs"
	files, err := os.ReadDir(glyph_path)
	if err != nil {
		fmt.Printf("Could not read content of folder %s\n", glyph_path)
		panic(err)
	}

	var glyph_list []string
	for _, file := range files {
		glyph_list = append(glyph_list, file.Name())
	}

	return glyph_list, nil
}
