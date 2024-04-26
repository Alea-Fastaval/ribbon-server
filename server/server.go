package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/user"
	"strconv"
)

type Server struct {
	socket net.Listener
}

func (server *Server) Start(socket_path string) {
	socket, err := net.Listen("unix", socket_path)
	if err != nil {
		panic(err)
	}
	server.socket = socket
	set_owner(socket_path)

	http.Handle("/", new(RequestHandler))

	go http.Serve(server.socket, nil)
	fmt.Println("Ribbon server listening on " + socket_path)
}

func (server Server) Stop() {
	server.socket.Close()
}

func set_owner(socket_path string) {
	www_data, err := user.Lookup("www-data")
	if err != nil {
		return
	}

	uid, err := strconv.Atoi(www_data.Uid)
	if err != nil {
		return
	}

	os.Chown(socket_path, uid, uid)
}
