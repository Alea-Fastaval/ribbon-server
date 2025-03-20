package connect

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/dreamspawn/ribbon-server/config"
)

var infosys_api string

func ConfigReady() {
	infosys_api = config.Get("infosys_api")
}

func GetUser(id string, pass string) map[string]string {
	result := make(map[string]string)

	response, err := http.PostForm(infosys_api, url.Values{"id": {id}, "pass": {pass}})
	if err != nil || response.StatusCode != 200 {
		log.Output(1, fmt.Sprintf("Login error: %+v\nResponse: %+v\n", err, response))
		return nil
	}

	var data map[string]interface{}
	json.NewDecoder(response.Body).Decode(&data)
	response.Body.Close()

	if status, ok := data["status"].(string); !ok || status != "success" {
		return nil
	}

	fields := []string{
		"name",
		"email",
	}

	for _, field := range fields {
		if value, ok := data[field].(string); ok {
			result[field] = value
		}
	}

	return result
}
