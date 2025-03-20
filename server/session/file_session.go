package session

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dreamspawn/ribbon-server/config"
	"github.com/dreamspawn/ribbon-server/user"
)

const session_name = "RibbonSession"
const session_prefix = "standard-session-"

var session_folder string

func ConfigReady() {
	session_folder = config.Get("resource_dir") + "sessions/"
}

type Session struct {
	id     string
	values map[string]string
}

func (session Session) Get(key string) string {
	return session.values[key]
}

func (session *Session) GetUser() *user.User {
	if session == nil {
		return nil
	}

	user_id_string := session.values["user-id"]
	if user_id_string == "" {
		return nil
	}

	user_id_uint, err := strconv.ParseUint(user_id_string, 10, 32)
	if err != nil {
		fmt.Printf("Could not parse user-id %s\n", user_id_string)
		return nil
	}

	return user.Load(uint(user_id_uint))
}

func (session Session) SetUser(user user.User) {
	session.values["user-id"] = strconv.FormatUint(uint64(user.ID), 10)
	write(session)
}

func Check(request http.Request) *Session {
	session_cookie, err := request.Cookie(session_name)
	if errors.Is(err, http.ErrNoCookie) {
		return nil
	}

	return load_from_file(session_cookie.Value)
}

func Start(writer http.ResponseWriter, request http.Request) *Session {
	// Get session cookie
	session_cookie, err := request.Cookie(session_name)
	if err == nil {
		session := load_from_file(session_cookie.Value)
		if session != nil {
			return session
		}
	}

	// Create new session cookie
	cookie := http.Cookie{
		Name:     session_name,
		Value:    create_new(),
		Path:     "/",
		MaxAge:   0,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(writer, &cookie)

	// Return new empty session
	return &Session{
		id:     cookie.Value,
		values: make(map[string]string),
	}
}

func (session Session) Save() {
	write(session)
}

func (session *Session) Delete(writer http.ResponseWriter) {
	// Delete session cookie
	cookie := http.Cookie{
		Name:     session_name,
		Value:    "undefined",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(writer, &cookie)

	// Delete session file
	filepath := session_folder + session.id
	err := os.Remove(filepath)
	if err != nil {
		fmt.Printf("Could not delete session file %s\n", filepath)
	}

	// Invalidate session variable
	session.id = ""
	session.values = nil
}

func create_new() string {
	file, err := os.CreateTemp(session_folder, session_prefix)
	if err != nil {
		fmt.Printf("Could not create new session file in folder %s\n", session_folder)
		panic(err)
	}

	file.Close()
	return filepath.Base(file.Name())
}

func load_from_file(id string) *Session {
	filepath := session_folder + id
	file, err := os.Open(filepath)

	// If file is gone, session is expired
	if err != nil {
		return nil
	}

	values := make(map[string]string)
	reader := bufio.NewReader(file)
	for {
		bytes, _, err := reader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Printf("Error reading session file %s\n", filepath)
			fmt.Printf("err: %v\n", err)
			return nil
		}

		tokens := strings.Split(string(bytes), " ")
		values[tokens[0]] = tokens[1]
	}

	os.Chtimes(filepath, time.Time{}, time.Now())

	return &Session{
		id:     id,
		values: values,
	}
}

func write(session Session) {
	content := ""

	for key, value := range session.values {
		content += key + " " + value + "\n"
	}

	err := os.WriteFile(session_folder+session.id, []byte(content), 0600)
	if err != nil {
		fmt.Printf("Could not write sesssion file: %s\n", session_folder+session.id)
	}
}
