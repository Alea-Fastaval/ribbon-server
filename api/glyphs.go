package api

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
)

func glyphsAPI(sub_path string, vars url.Values, request http.Request) (any, error) {
	glyph_path := resource_dir + "public/glyphs/"

	file_system := os.DirFS(glyph_path)
	matches, err := fs.Glob(file_system, "*.svg")

	if err != nil {
		fmt.Printf("Could not read content of folder %s\n", glyph_path)
		panic(err)
	}

	return matches, nil
}
