package redeye

import (
	"log"
	"net/http"
)

type WebServer struct {
	Addr	string
	*http.ServeMux
}

func NewWebServer(addr string) *WebServer {
	return &WebServer{
		Addr: addr,
		ServeMux: http.NewServeMux(),
	}
}

func (srv *WebServer) Listen() {
	log.Println("Capturing. Point your browser to " + srv.Addr)
	log.Fatal(http.ListenAndServe(srv.Addr, srv.ServeMux))	
}

