package render

import (
	"fmt"
	"math"

	"codeberg.org/go-pdf/fpdf"
	"github.com/dreamspawn/ribbon-server/database"
	"github.com/dreamspawn/ribbon-server/user"
)

var ribbon_width = 25.0
var ribbon_height = 8.4

func UserCollection(user user.User, pdf *fpdf.Fpdf) error {
	collection, err := database.GetOrders(user.ID)
	if err != nil {
		fmt.Printf("Could not get orders for user %d", user.ID)
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
	last_count := int64(len(ribbons)) % columns

	pw, ph := pdf.GetPageSize()
	_, _, _, mb := pdf.GetMargins()

	pc := pw / 2.0
	x_start := pc - ribbon_width*(float64(columns)/2.0)

	if pdf.GetY()+(float64(rows)*ribbon_height)+10.0 >= ph-mb {
		pdf.AddPage()
	}

	user_text := fmt.Sprintf("%s, ID: %d", user.Name, user.ParticipantID)
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

		// Check for last column
		flow := false
		last_col := columns - 1
		if row == rows-1 && last_count != 0 {
			last_col = last_count - 1
		}
		if col == last_col {
			flow = true
			row++
		}

		pdf.ImageOptions(ribbon_png, x, y, ribbon_width, 0, flow, options, 0, "")
		col = (col + 1) % columns

		// Starting last row
		if row == rows-1 && col == 0 {
			if last_count != 0 {
				x_start = pc - ribbon_width*(float64(last_count)/2.0)
			}
		}
	}

	// Add bottom margin
	pdf.SetY(pdf.GetY() + 10)

	return nil
}
