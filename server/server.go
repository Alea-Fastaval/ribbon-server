package server

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/dreamspawn/ribbon-server/util"
)

type Server struct {
	socket net.Listener
}

func (server *Server) Start(socket_path string) {
	socket, err := net.Listen("unix", socket_path)
	if err != nil {
		fmt.Printf("\nCouldn't bind socket %s\n", socket_path)
		if strings.Contains(err.Error(), "permission denied") {
			fmt.Print("\x1b[41mYou may need to run the server as root\x1b[0m\n\n")
		}
		panic(err)
	}
	server.socket = socket
	util.SetOwner(socket_path)

	http.Handle("/", new(RequestHandler))

	go http.Serve(server.socket, nil)
	fmt.Println("Ribbon server listening on " + socket_path)
}

func (server Server) Stop() {
	server.socket.Close()
}
