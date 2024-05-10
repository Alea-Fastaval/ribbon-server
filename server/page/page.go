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
	title_header := header{
		"Type":  "title",
		"Value": page_title,
	}
	page.Headers = append(page.Headers, title_header)
}

func (page *Page) AddCSS(path string) {
	path = "/public/css/" + path
	info, err := os.Stat(resource_dir + path)
	if err == nil {
		mod_time := fmt.Sprintf("?v=%d", info.ModTime().Unix())
		path += mod_time
	} else {
		fmt.Printf("There was an error loading file info for file: %s\n%+v", resource_dir+path, err)
	}

	css_header := header{
		"Type":  "css",
		"Value": path,
	}
	page.Headers = append(page.Headers, css_header)
}

func (page *Page) AddJS(path string) {
	path = "/public/js/" + path
	path += get_file_version(path)

	js_header := header{
		"Type":  "js",
		"Value": path,
	}
	page.Headers = append(page.Headers, js_header)
}

func get_file_version(path string) string {
	info, err := os.Stat(resource_dir + path)

	if err != nil {
		fmt.Printf("There was an error loading file info for file: %s\n%+v", resource_dir+path, err)
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
