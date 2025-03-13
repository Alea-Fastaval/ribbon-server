package user

import (
	"net/url"

	"github.com/dreamspawn/ribbon-server/connect"
)

type User struct {
	ID      uint
	IsAdmin bool
	Name    string
}

var admin = User{
	1, true, "Admin",
}

var test = User{
	2, false, "Test",
}

func Load(id uint) *User {
	if id == 1 {
		return &admin
	}
	return &test
}

func TryLogin(vars url.Values) *User {
	if vars["user-name"][0] == "Admin" {
		return &admin
	}

	result := connect.GetUser(vars["user-name"][0], vars["password"][0])

	if result == nil || result["status"] == "error" {
		return nil
	}

	return &User{
		ID:      2,
		IsAdmin: false,
		Name:    result["name"],
	}
}

func (user *User) CheckAccess(base, sub, method string) bool {
	// TODO return true for any endpoint allowed for even anonymous users

	if user == nil {
		return false
	}

	if user.IsAdmin {
		return true
	}

	// TODO maybe more detailed user permissions
	return false
}
