package api

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dreamspawn/ribbon-server/config"
	"github.com/dreamspawn/ribbon-server/server/session"
	"github.com/dreamspawn/ribbon-server/user"
	"github.com/golang-jwt/jwt/v5"
)

var resource_dir string
var hmac_key []byte

func ConfigReady() {
	resource_dir = config.Get("resource_dir")
	set_hmac_key(resource_dir + "keys/hmac.key")
	export_config_ready()
}

var endpoints = map[string]func(string, url.Values, http.Request) (any, error){
	"categories":   categoriesAPI,
	"translations": translationsAPI,
	"glyphs":       glyphsAPI,
	"ribbons":      ribbonsAPI,
	"orders":       ordersAPI,
	"export":       exportAPI,
	"options":      optionsAPI,
}

func Handle(endpoint string, vars url.Values, request http.Request, writer http.ResponseWriter) {
	base, sub_path, _ := strings.Cut(endpoint, "/")
	var current_user *user.User
	//-----------------------------------------------------
	// Error handling
	//-----------------------------------------------------
	defer func() {
		result := recover()
		if result == nil {
			return
		}

		var message string
		var err error

		if array, ok := result.([]interface{}); ok {
			message = array[0].(string)
			err = array[1].(error)
		} else {
			// Log non-API errors
			log.Output(1, fmt.Sprintf("API error handler: %+v\n", result))
		}

		var message_string, padding string

		message_array := strings.Split(message, "\n")
		for _, line := range message_array {
			message_string += padding
			message_string += fmt.Sprintf(`"%s"`, line)
			padding = ",\n"
		}

		padding = ""
		error_string := ""
		error_array := strings.Split(fmt.Sprintf("%v", err), "\n")
		for _, line := range error_array {
			error_string += padding
			error_string += fmt.Sprintf(`"%s"`, line)
			padding = ",\n"
		}

		response := fmt.Sprintf(
			`{
				"status" : "error",
				"message" : [%s],
				"error": [%s]
			}`, message_string, error_string,
		)

		writer.WriteHeader(400)
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		io.WriteString(writer, response)
	}()

	//-----------------------------------------------------
	// Authenticate user
	//-----------------------------------------------------
	auth_header := ""
	if header, ok := request.Header["Authorization"]; ok {
		auth_header = header[0]
	}
	if auth_header != "" {
		// Get user from uath header
		auth_schema, auth_payload, _ := strings.Cut(auth_header, " ")
		switch auth_schema {
		case "Bearer":
			token, err := jwt.Parse(auth_payload, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}

				return hmac_key, nil
			})
			if err != nil {
				api_error("Could not parse auth payload", err)
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				var exp int64
				if exp_string, ok := claims["exp"].(string); ok {
					exp, err = strconv.ParseInt(exp_string, 10, 64)
					if err != nil {
						api_error("Could not parse JWT expiration timestamp", err)
					}
				} else {
					api_error("Missing or malformed expiration in JWT", nil)
				}

				if time.Now().Unix() > exp {
					api_error("JWT has expiredT", nil)
				}

				if uid_string, ok := claims["uid"].(string); ok {
					uid, err := strconv.ParseUint(uid_string, 10, 64)
					if err != nil {
						api_error(fmt.Sprintf("Failed to parse user id: %d", uid), err)
					}

					current_user = user.Load(uint(uid))
					if current_user == nil {
						api_error("Invalid user id", nil)
					}
				} else {
					api_error("Missing or malformed user id", nil)
				}

			} else {
				api_error("No claims in JWT", err)
			}

		default:
			api_error(fmt.Sprintf("Auth schema %s is not implemented", auth_schema), nil)
		}
	} else {
		// Get user from session
		session := session.Check(request)
		current_user = session.GetUser()
	}

	if !current_user.CheckAccess(base, sub_path, request.Method) {
		api_error(fmt.Sprintf("User %s does not have access to %s at %s", current_user.GetName(), request.Method, endpoint), nil)
	}

	//-----------------------------------------------------
	// Perform request
	//-----------------------------------------------------
	var data any
	var err error

	if function, ok := endpoints[base]; ok {
		data, err = function(sub_path, vars, request)
	} else {
		api_error(fmt.Sprintf("Endpoint %s is not defined", base), nil)
	}

	if err != nil {
		if message, ok := data.(string); ok {
			api_error(message, err)
		}
		api_error("Could not perform API request", err)
	}

	//-----------------------------------------------------
	// Format output
	//-----------------------------------------------------

	// Special data format
	if data_map, ok := data.(map[string]string); ok {
		if content_type, ok := data_map["content_type"]; ok {
			writer.Header().Set("Content-Type", content_type+"; charset=utf-8")
		}

		if output, ok := data_map["output"]; ok {
			io.WriteString(writer, output)
			return
		}

		if file_path, ok := data_map["file_path"]; ok {
			if filename, ok := data_map["file_name"]; ok {
				writer.Header().Set("Content-Disposition", "attachment; filename="+filename)
			}
			http.ServeFile(writer, &request, file_path)
		}
	}

	// JSON output
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	json_bytes, err := json.Marshal(data)
	if err != nil {
		api_error("Error parsing output", err)
	}

	var pretty bytes.Buffer
	err = json.Indent(&pretty, json_bytes, "", "  ")
	if err != nil {
		api_error("Error formating output", err)
	}

	io.WriteString(writer, pretty.String())
}

func api_error(message string, err error) {
	log.Output(2, "API ERROR MSG: "+message)

	if err != nil {
		log.Output(2, "API ERROR:"+err.Error())
	} else {
		err = fmt.Errorf("nil")
	}

	panic([]interface{}{
		message,
		err,
	})
}

func set_hmac_key(path string) {
	hmac_key_content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		hmac_key_content = make([]byte, 16)
		rand.Read(hmac_key_content)
		err = os.WriteFile(path, hmac_key_content, 0600)
	}

	if err != nil {
		log.Output(1, fmt.Sprintf("Could not open HMAC key file: %s", path))
		panic(err)
	}

	hmac_key = hmac_key_content
}
