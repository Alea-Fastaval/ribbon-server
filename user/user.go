package user

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/dreamspawn/ribbon-server/config"
	"github.com/dreamspawn/ribbon-server/connect"
	"github.com/dreamspawn/ribbon-server/database"
)

var ErrNoUserFound = errors.New("no user found")

type User struct {
	ID      uint
	IsAdmin bool
	Name    string
}

var admin = User{
	0, true, "Admin",
}

func (user *User) GetName() string {
	if user == nil {
		return "<anonymous>"
	}
	return user.Name
}

func (user *User) CheckAccess(base, sub, method string) bool {
	// TODO return true for any endpoint allowed for even anonymous users

	if user == nil {
		return false
	}

	if user.IsAdmin {
		return true
	}

	if base == "orders" {
		return true
	}

	if method == "GET" {
		return true
	}

	// TODO maybe more detailed user permissions
	return false
}

func Load(id uint) *User {
	if id == 0 {
		return &admin
	}

	user, err := loadWithUserIDFromDB(id)
	if errors.Is(err, ErrNoUserFound) {
		return nil
	}
	if err != nil {
		log.Output(2, fmt.Sprintf("Error loading user with ID: %d from DB:\n %+v\n", id, err))
		return nil
	}

	return user
}

func TryLogin(vars url.Values) *User {
	if vars["user-name"][0] == "Admin" {
		if vars["password"][0] != "" && vars["password"][0] == config.Get("admin_pass") {
			return &admin
		} else {
			return nil
		}
	}

	participant_id := vars["user-name"][0]
	result := connect.GetUser(participant_id, vars["password"][0])

	if result == nil || result["status"] == "error" {
		return nil
	}

	this_year := time.Now().Format("2006")
	user, err := loadParticipantFromDB(participant_id, this_year)
	if err != nil && !errors.Is(err, ErrNoUserFound) {
		log.Output(2, fmt.Sprintf("Error loading user from DB: %+v\n", err))
		return nil
	}

	if user == nil {
		user = createInDB(
			participant_id,
			this_year,
			result["name"],
			result["email"],
		)
	}

	return user
}

func loadWithUserIDFromDB(uid uint) (*User, error) {
	query := "SELECT * FROM users WHERE id = ?"
	return queryLoadFromDB(query, []any{uid})
}

func loadParticipantFromDB(pid, year string) (*User, error) {
	query := "SELECT * FROM users WHERE participant_id = ? AND year = ?"
	return queryLoadFromDB(query, []any{pid, year})
}

func queryLoadFromDB(query string, args []any) (*User, error) {
	result, err := database.Query(query, args)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, ErrNoUserFound
	}

	return &User{
		ID:      uint(result[0]["id"].(int64)),
		Name:    result[0]["name"].(string),
		IsAdmin: false,
	}, nil
}

func createInDB(pid, year, name, email string) *User {
	query := "INSERT INTO users (participant_id, year, name, email, status) VALUES(?,?,?,?,'open')"
	result, err := database.Exec(query, []any{pid, year, name, email})
	if err != nil {
		log.Output(2, fmt.Sprintf("Error creating new DB user: %+v\n", err))
		return nil
	}

	id, _ := result.LastInsertId()

	return &User{
		ID:      uint(id),
		Name:    name,
		IsAdmin: false,
	}
}
