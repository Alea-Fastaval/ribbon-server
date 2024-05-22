package user

import "net/url"

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
	return &test
}
