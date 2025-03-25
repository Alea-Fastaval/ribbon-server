package page

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/dreamspawn/ribbon-server/config"
)

type header map[string]string

type Page struct {
	Headers               []header
	Content               string
	Lang                  string
	Explanation_link_text string
	Long_explanation      string
}

var resource_dir string

func ConfigReady() {
	resource_dir = config.Get("resource_dir")
}

func (page *Page) SetExplanation(link, text string) {
	page.Explanation_link_text = link
	page.Long_explanation = text
}

func (page *Page) AddTitle(page_title string) {
	page.Headers = append(page.Headers, header{
		"Type":  "title",
		"Value": page_title,
	})
}

func (page *Page) AddCSS(path string) {
	external_path := "public/css/" + path
	local_path := resource_dir + external_path
	info, err := os.Stat(local_path)

	if err != nil {
		fmt.Printf("There was an error loading file info for file: %s\n%+v\n", resource_dir+path, err)
	}

	if info.IsDir() {
		file_system := os.DirFS(resource_dir)
		fs.WalkDir(file_system, external_path, func(path string, file fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if file.IsDir() {
				return nil
			}

			info, err := file.Info()
			if err != nil {
				return err
			}

			add_css_header(page, info, path)
			return nil
		})
		return
	}

	add_css_header(page, info, external_path)
}

func add_css_header(page *Page, info os.FileInfo, path string) {
	path += fmt.Sprintf("?v=%d", info.ModTime().Unix())
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

func (page *Page) Prepend(content string) {
	page.Content = content + page.Content
}
