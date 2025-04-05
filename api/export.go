package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"codeberg.org/go-pdf/fpdf"
	"github.com/dreamspawn/ribbon-server/render"
	"github.com/dreamspawn/ribbon-server/user"
)

var export_file = "export.pdf"
var export_folder string

func export_config_ready() {
	export_folder = resource_dir + "tmp/"
}

// /api/export
func exportAPI(sub_path string, vars url.Values, request http.Request) (any, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	this_year, _ := strconv.ParseInt(time.Now().Format("2006"), 0, 0)
	user_list, err := user.GetAllFromYear(int(this_year))
	if err != nil {
		api_error(fmt.Sprintf("could not load users from year %d", this_year), err)
	}

	for _, user := range user_list {
		err = render.UserCollection(user.ID, user.Name, pdf)
		if err != nil {
			api_error(fmt.Sprintf("failed to render user collection for user %d", 1), err)
		}
	}

	err = pdf.OutputFileAndClose(export_folder + export_file)

	if err != nil {
		api_error("failed to create PDF", err)
	}

	return map[string]string{
		"status":       "success",
		"content_type": "application/pdf",
		//"file_name":    export_file,
		"file_path": export_folder + export_file,
		//"output": string(ribbon_png),
	}, nil

}
