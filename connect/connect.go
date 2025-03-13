package connect

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func GetUser(id string, pass string) map[string]string {
	result := make(map[string]string)

	response, err := http.PostForm("https://infosys.fastaval.dk/api/ribbon/login", url.Values{"id": {id}, "pass": {pass}})
	if err != nil || response.StatusCode != 200 {
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

	// fmt.Printf("Status: %+v\n", response.Status)
	// fmt.Printf("Headers: %+v\n", response.Header)
	// fmt.Printf("GetUser data: %+v\n", data)

	for _, field := range fields {
		if value, ok := data[field].(string); ok {
			result[field] = value
		}
	}

	// fmt.Printf("GetUser result: %+v\n", result)

	return result
}
