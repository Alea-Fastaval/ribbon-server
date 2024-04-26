package server

import (
	"io"
	"net/http"
)

type RequestHandler struct {
}

func (handler RequestHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	io.WriteString(writer, "Fastaval Ribbon Machine<br>")
	//fmt.Fprintf(writer, "%+v", request)
}
