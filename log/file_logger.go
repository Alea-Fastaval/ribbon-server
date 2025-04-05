package log

import (
	"fmt"
	"log"
	"os"

	"github.com/dreamspawn/ribbon-server/config"
)

func Init() {
	logfile_path := config.Get("resource_dir") + "log.txt"
	logfile, err := os.OpenFile(logfile_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Could not open log file: %s\n", logfile_path)
		return
	}

	log.SetOutput(logfile)
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
}
