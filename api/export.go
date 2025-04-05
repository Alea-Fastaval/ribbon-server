package api

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"codeberg.org/go-pdf/fpdf"
	"github.com/dreamspawn/ribbon-server/render"
	"github.com/dreamspawn/ribbon-server/user"
)

var export_file = "export.pdf"
var public_export_folder = "public/export/"
var public_folder_url = "/" + public_export_folder
var export_folder string
var public_folder string

func export_config_ready() {
	export_folder = resource_dir + "tmp/"
	public_folder = resource_dir + public_export_folder
}

// /api/export
func exportAPI(sub_path string, vars url.Values, request http.Request) (any, error) {
	pdf := fpdf.New("P", "mm", "A4", resource_dir+"font/")
	pdf.AddUTF8Font("Montserrat", "", "montserrat_regular.ttf")
	pdf.SetFont("Montserrat", "", 12)
	pdf.AddPage()

	this_year, _ := strconv.ParseInt(time.Now().Format("2006"), 0, 0)
	user_list, err := user.GetAllFromYear(int(this_year))
	if err != nil {
		api_error(fmt.Sprintf("could not load users from year %d", this_year), err)
	}

	for _, user := range user_list {
		err = render.UserCollection(user, pdf)
		if err != nil {
			api_error(fmt.Sprintf("failed to render user collection for user %d", 1), err)
		}
		pdf.SetY(pdf.GetY() + 10)
	}

	err = pdf.OutputFileAndClose(export_folder + export_file)

	if err != nil {
		api_error("failed to create PDF", err)
	}

	err = os.Rename(export_folder+export_file, public_folder+export_file)
	if err != nil {
		api_error("filed to move export file", err)
	}

	return map[string]string{
		"status":        "success",
		"download_file": public_folder_url + export_file,
	}, nil
}
