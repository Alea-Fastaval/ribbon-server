package render

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/dreamspawn/ribbon-server/config"
	"github.com/dreamspawn/ribbon-server/database"
)

var tmpl_folder string
var temp_folder string

func ConfigReady() {
	resource_dir := config.Get("resource_dir")
	tmpl_folder = resource_dir + "templates/"
	temp_folder = resource_dir + "tmp/ribbons/"
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

func Ribbon(ribbon database.Ribbon, category database.Category, glyph database.Glyph, years, leader, second uint64) string {
	svg_data := map[string]interface{}{
		"background": category.Background,
	}

	// Render stripes
	seniority_tmpl := FindYearTemplate(int(years), "seniority")

	if seniority_tmpl != nil {
		seniority_data := map[string]interface{}{
			"fill_color": category.Stripes,
			"background": category.Background,
		}
		svg_data["seniority"] = TemplateString(seniority_tmpl, seniority_data)
	}

	// Render wings
	wing_years := uint64(0)
	wing_color := ""
	if second > 0 {
		wing_years = second
		wing_color = category.Wing2
	}
	if leader > 0 {
		wing_years += leader
		wing_color = category.Wing1
	}

	if wings_rule, ok := ribbon.Special["always_wings"]; ok {
		wing_info := strings.Split(wings_rule, ",")
		if wing_info[0] == "1" {
			wing_color = category.Wing1
		} else {
			wing_color = category.Wing2
		}
		wing_years, _ = strconv.ParseUint(wing_info[1], 0, 0)
	}

	if wing_years > 0 {
		wing_tmpl := FindYearTemplate(int(wing_years), "wings")

		if wing_tmpl != nil {
			wing_data := map[string]interface{}{
				"foreground": wing_color,
			}
			svg_data["wings"] = TemplateString(wing_tmpl, wing_data)
		}
	}

	// Render glyph
	glyph_data := map[string]string{
		"foreground": category.Glyph,
		"background": category.Background,
	}

	if color, ok := ribbon.Special["glyph_color"]; ok {
		glyph_data["foreground"] = color
	}

	glyph_template := LoadTemplate(glyph.File)
	svg_data["glyph"] = TemplateString(glyph_template, glyph_data)

	// Render ribbon
	svg_tmpl := LoadTemplate("full-ribbon.tmpl")
	return TemplateString(svg_tmpl, svg_data)
}

func RibbonFromOrder(order_id uint) (string, error) {
	result, err := database.GetOrderByID(order_id)
	if err != nil {
		return "", err
	}
	order := result["order"].(map[string]any)

	// Load ribbon
	ribbon, err := database.GetRibbon(uint(order["ribbon_id"].(int64)))
	if err != nil {
		return "", err
	}

	// Load category
	category, err := ribbon.GetCategory()
	if err != nil {
		return "", err
	}

	second := uint64(order["second"].(int64))
	leader := uint64(order["leader"].(int64))

	years := uint64(order["grunt"].(int64)) + second + leader

	// Load glyph
	glyph, err := database.GetGlyph(ribbon.Glyph)
	if err != nil {
		return "", err
	}

	return Ribbon(*ribbon, *category, *glyph, years, leader, second), nil
}

func PNGFromOrder(order_id uint) (string, error) {
	tmp_svg := temp_folder + fmt.Sprintf("ribbon%d.svg", order_id)
	tmp_png := temp_folder + fmt.Sprintf("ribbon%d.png", order_id)

	existing_file, err := os.Open(tmp_png)
	if err == nil {
		stats, err := existing_file.Stat()
		if err == nil {
			if time.Since(stats.ModTime()).Minutes() < 10 {
				return tmp_png, nil
			}
		}
	}

	svg_data, err := RibbonFromOrder(order_id)
	if err != nil {
		return "", err
	}

	svg_file, err := os.Create(tmp_svg)
	if err != nil {
		fmt.Printf("Error creating file %s\n", tmp_svg)
		return "", err
	}
	svg_file.Write([]byte(svg_data))
	svg_file.Close()

	var stderr bytes.Buffer

	cmd := exec.Command("/usr/bin/inkscape", "-z", "-w 250", "-h 84", "--export-png="+tmp_png, tmp_svg)
	cmd.Stderr = &stderr

	if e := cmd.Run(); e != nil {
		err = fmt.Errorf("%s\nSTDERR:\n%s", e.Error(), stderr.String())
		return "", err
	}

	return tmp_png, nil
}
