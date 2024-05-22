package util

import (
	"os"
	"os/user"
	"strconv"
)

func SetOwner(path string) {
	www_data, err := user.Lookup("www-data")
	if err != nil {
		return
	}

	uid, err := strconv.Atoi(www_data.Uid)
	if err != nil {
		return
	}

	os.Chown(path, uid, uid)
}
