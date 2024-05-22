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

	"github.com/dreamspawn/ribbon-server/config"
	"github.com/dreamspawn/ribbon-server/user"
	"github.com/dreamspawn/ribbon-server/util"
)

const session_name = "RibbonSession"
const session_prefix = ""

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

func (session Session) GetUser() *user.User {
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

func Open(writer http.ResponseWriter, request http.Request) Session {
	session := Session{}
	session.values = make(map[string]string)

	session_cookie, err := request.Cookie(session_name)
	if errors.Is(err, http.ErrNoCookie) {
		session.id = create_new()

		cookie := http.Cookie{
			Name:     session_name,
			Value:    session.id,
			MaxAge:   0,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}

		http.SetCookie(writer, &cookie)
		return session
	}

	session.id = session_cookie.Value
	load_from_file(&session)
	return session
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

func load_from_file(session *Session) {
	filepath := session_folder + session.id
	file, err := os.Open(filepath)

	// Create new session file if old one is gone
	if err != nil {
		file, err = os.Create(filepath)
		if err != nil {
			fmt.Printf("Could not create new session file %s\n", filepath)
			panic(err)
		}

		file.Close()
		util.SetOwner(filepath)
		return
	}

	reader := bufio.NewReader(file)
	for {
		bytes, _, err := reader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Printf("Error reading session file %s\n", filepath)
			fmt.Printf("err: %v\n", err)
			return
		}

		tokens := strings.Split(string(bytes), " ")
		session.values[tokens[0]] = tokens[1]
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
