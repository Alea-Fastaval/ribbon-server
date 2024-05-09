package api

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/dreamspawn/ribbon-server/database"
)

func Handle(endpoint string, query url.Values) string {
	data, err := database.Query("SELECT * FROM test", nil)
	if err != nil {
		panic(err)
	}

	json_bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	var pretty bytes.Buffer
	err = json.Indent(&pretty, json_bytes, "", "  ")
	if err != nil {
		panic(err)
	}

	return pretty.String()
}
