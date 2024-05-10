package page

import (
	"fmt"
	"os"

	"github.com/dreamspawn/ribbon-server/config"
)

type header map[string]string

type Page struct {
	Headers []header
	Content string
	Lang    string
}

var resource_dir string

func ConfigReady() {
	resource_dir = config.Get("resource_dir")
}

func (page *Page) AddTitle(page_title string) {
	page.Headers = append(page.Headers, header{
		"Type":  "title",
		"Value": page_title,
	})
}

func (page *Page) AddCSS(path string) {
	path = "public/css/" + path
	path += get_file_version(path)

	page.Headers = append(page.Headers, header{
		"Type":  "css",
		"Value": path,
	})
}

func (page *Page) AddJS(path string) {
	path = "public/js/" + path
	path += get_file_version(path)

	page.Headers = append(page.Headers, header{
		"Type":  "js",
		"Value": path,
	})
}

func get_file_version(path string) string {
	info, err := os.Stat(resource_dir + path)

	if err != nil {
		fmt.Printf("There was an error loading file info for file: %s\n%+v\n", resource_dir+path, err)
		return ""
	}

	mod_time := fmt.Sprintf("?v=%d", info.ModTime().Unix())
	return mod_time
}

func (page Page) GetHeaders() []header {
	return page.Headers
}

func (page *Page) SetContent(content string) {
	page.Content = content
}
