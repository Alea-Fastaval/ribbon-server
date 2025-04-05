package render

import (
	"fmt"
	"math"

	"codeberg.org/go-pdf/fpdf"
	"github.com/dreamspawn/ribbon-server/database"
)

var ribbon_width = 25.0
var ribbon_height = 10.0

func UserCollection(uid uint, name string, pdf *fpdf.Fpdf) error {
	collection, err := database.GetOrders(uid)
	if err != nil {
		fmt.Printf("Could not get orders for user %d", uid)
		return err
	}

	ribbons := collection["list"].(map[uint]map[string]any)
	if len(ribbons) == 0 {
		return nil
	}

	ribbons_ordered := make([]map[string]any, len(ribbons))
	for _, ribbon := range ribbons {
		ribbons_ordered[ribbon["position"].(int64)] = ribbon
	}

	settings := collection["settings"].(map[string]any)
	columns := settings["columns"].(int64)
	rows := int64(math.Ceil(float64(len(ribbons)) / float64(columns)))

	pw, ph := pdf.GetPageSize()
	pc := pw / 2
	x_start := pc - ribbon_width*(float64(columns)/2.0)

	if pdf.GetY()+(float64(rows)*ribbon_height)+10 > ph {
		pdf.AddPage()
	}

	user_text := fmt.Sprintf("%s, ID: %d", name, uid)
	pdf.CellFormat(0, 10, user_text, "", 1, "C", false, 0, "")

	options := fpdf.ImageOptions{
		ImageType:             "PNG",
		ReadDpi:               false,
		AllowNegativePosition: false,
	}

	col := int64(0)
	row := int64(0)
	for _, order := range ribbons_ordered {
		ribbon_png, err := PNGFromOrder(uint(order["id"].(int64)))
		if err != nil {
			fmt.Printf("Error rendering PNG from order %d:\n%v\n", order["id"], err)
			return err
		}

		y := pdf.GetY()
		x := x_start + float64(col)*ribbon_width
		flow := false
		if col == columns-1 {
			flow = true
			row++
		}
		pdf.ImageOptions(ribbon_png, x, y, 25, 0, flow, options, 0, "")
		col = (col + 1) % columns
		// Stating last row
		if row == rows-1 && col == 0 {
			last_count := len(ribbons) % int(columns)
			if last_count != 0 {
				x_start = pc - ribbon_width*(float64(last_count)/2.0)
			}
		}
	}

	return nil
}
