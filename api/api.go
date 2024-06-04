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
	set_hmac_key(resource_dir + "/keys/hmac.key")
}

var endpoints = map[string]func(string, url.Values, http.Request) (any, error){
	"categories":   categoriesAPI,
	"translations": translationsAPI,
	"glyphs":       glyphsAPI,
}

func Handle(endpoint string, vars url.Values, request http.Request, writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	base, sub_path, _ := strings.Cut(endpoint, "/")
	var current_user *user.User

	//-----------------------------------------------------
	// Authenticate user
	//-----------------------------------------------------
	auth_header := request.Header["Authorization"][0]
	if auth_header != "" {
		// Get user from uath header
		auth_schema, auth_payload, _ := strings.Cut(auth_header, " ")
		switch auth_schema {
		case "Bearer":
			token, err := jwt.Parse(auth_payload, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				return hmac_key, nil
			})
			if err != nil {
				api_error(writer, "Could not parse auth payload", err)
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				var exp int64
				if exp_string, ok := claims["exp"].(string); ok {
					exp, err = strconv.ParseInt(exp_string, 10, 64)
					if err != nil {
						api_error(writer, "Could not parse JWT expiration timestamp", err)
					}
				} else {
					api_error(writer, "Missing or malformed expiration in JWT", nil)
				}

				if time.Now().Unix() > exp {
					api_error(writer, "JWT has expiredT", nil)
				}

				if uid_string, ok := claims["uid"].(string); ok {
					uid, err := strconv.ParseUint(uid_string, 10, 64)
					if err != nil {
						api_error(writer, fmt.Sprintf("Failed to parse user id: %d", uid), err)
					}

					current_user = user.Load(uint(uid))
					if current_user == nil {
						api_error(writer, "Invalid user id", nil)
					}
				} else {
					api_error(writer, "Missing or malformed user id", nil)
				}

			} else {
				api_error(writer, "No claims in JWT", err)
			}

		default:
			api_error(writer, fmt.Sprintf("Auth schema %s is not implemented", auth_schema), nil)
		}
	} else {
		// Get user from session
		session := session.Open(writer, request)
		current_user = session.GetUser()
	}

	if !current_user.CheckAccess(base, sub_path, request.Method) {
		api_error(writer, fmt.Sprintf("User %s does not have access to %s at %s", current_user.Name, request.Method, endpoint), nil)
	}

	//-----------------------------------------------------
	// Perform request
	//-----------------------------------------------------
	var data any
	var err error

	if function, ok := endpoints[base]; ok {
		data, err = function(sub_path, vars, request)
	} else {
		api_error(writer, fmt.Sprintf("Endpoint %s is not defined", base), nil)
	}

	if err != nil {
		api_error(writer, "Could not perform API request", err)
	}

	//-----------------------------------------------------
	// Format output
	//-----------------------------------------------------
	json_bytes, err := json.Marshal(data)
	if err != nil {
		api_error(writer, "Error parsing output", err)
	}

	var pretty bytes.Buffer
	err = json.Indent(&pretty, json_bytes, "", "  ")
	if err != nil {
		api_error(writer, "Error formating output", err)
	}

	io.WriteString(writer, pretty.String())
}

func api_error(writer http.ResponseWriter, message string, err error) {
	response := fmt.Sprintf(
		`{
      "status" : "error",
      "message" : "%s",
      "error": %+v
    }`, message, err,
	)
	io.WriteString(writer, response)
	log.Fatal(err)
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
