package api

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"codeberg.org/go-pdf/fpdf"
	"github.com/dreamspawn/ribbon-server/database"
	"github.com/dreamspawn/ribbon-server/render"
	"github.com/dreamspawn/ribbon-server/user"
)

var export_file = "export.pdf"
var public_export_folder = "public/export/"

// var public_folder_url = "/" + public_export_folder
var export_folder string
var public_folder string

var export_count uint
var export_time time.Time

func export_config_ready() {
	export_folder = resource_dir + "tmp/"
	public_folder = resource_dir + public_export_folder
}

// /api/export
func exportAPI(sub_path string, vars url.Values, request http.Request) (any, error) {
	//--------------------------------------------------------------------------------------------------------------------
	// Get time of last export
	//--------------------------------------------------------------------------------------------------------------------
	if strings.HasPrefix(sub_path, "time") {
		file, err := os.Open(public_folder + export_file)
		if err != nil {
			api_error("could not open export file", err)
		}
		stats, err := file.Stat()
		if err != nil {
			api_error("could not stat export file", err)
		}
		return map[string]any{
			"status":           "success",
			"file_age_minutes": time.Since(stats.ModTime()).Minutes(),
		}, nil
	}

	//--------------------------------------------------------------------------------------------------------------------
	// Get export progress
	//--------------------------------------------------------------------------------------------------------------------
	if strings.HasPrefix(sub_path, "progress") {
		if export_count == 0 {
			return map[string]any{
				"total":       export_count,
				"processed":   0,
				"export_time": export_time.Unix(),
			}, nil
		}

		files, err := os.ReadDir(export_folder + "ribbons")
		if err != nil {
			api_error("could not read export folder", err)
		}

		count := 0
		for _, file := range files {
			info, _ := file.Info()
			if info.Mode().IsRegular() && filepath.Ext(file.Name()) == ".png" && info.ModTime().After(export_time) {
				count++
			}
		}

		return map[string]any{
			"total":       export_count,
			"processed":   count,
			"export_time": export_time.Unix(),
		}, nil
	}

	//--------------------------------------------------------------------------------------------------------------------
	// Create new export
	//--------------------------------------------------------------------------------------------------------------------
	if export_count != 0 {
		api_error(
			"could not start export",
			fmt.Errorf("export already running"),
		)
	}

	this_year, _ := strconv.ParseInt(time.Now().Format("2006"), 0, 0)

	// Make note of starting export with time and how many orders
	var err error
	export_count, err = database.GetOrderCount(uint(this_year))
	if err != nil {
		api_error(fmt.Sprintf("could not get order count for year %d", this_year), err)
	}
	export_time = time.Now()

	go do_export(int(this_year))

	return map[string]any{
		"status":      "success",
		"start time":  export_time.Unix(),
		"order count": export_count,
	}, nil
}

func do_export(this_year int) {
	defer func() {
		export_count = 0
		log.Output(1, "Resetting export counter")
	}()

	pdf := fpdf.New("P", "mm", "A4", resource_dir+"font/")
	pdf.AddUTF8Font("Montserrat", "", "montserrat_regular.ttf")
	pdf.SetFont("Montserrat", "", 12)
	pdf.AddPage()

	user_list, err := user.GetAllFromYear(this_year)
	if err != nil {
		log.Output(1, fmt.Sprintf("Could not load users from year %d\n%v\n", this_year, err))
		return
	}

	for _, user := range user_list {
		err = render.UserCollection(user, pdf)
		if err != nil {
			log.Output(1, fmt.Sprintf("Failed to render user collection for user %d\n%v\n", user.ID, err))
			return
		}
	}

	err = pdf.OutputFileAndClose(export_folder + export_file)

	if err != nil {
		log.Output(1, fmt.Sprintf("Failed to create PDF: %s\n%+v\n", export_folder+export_file, err))
		return
	}

	err = os.Rename(export_folder+export_file, public_folder+export_file)
	if err != nil {
		log.Output(1, fmt.Sprintf("Failed to move PDF\nFrom: %s\nTo:%s\n%+v\n", export_folder+export_file, public_folder+export_file, err))
		return
	}

	log.Output(1, "Finished PDF export")
}
