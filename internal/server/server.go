package server

import (
	"net/http"

	"github.com/chavacava/tuilx/internal/tile/source"
	"github.com/gorilla/mux"
)

// Server is a tile server over HTTP
type Server struct {
	router *mux.Router
	source source.Source
}

// New yields a fresh new Server
func New(source source.Source) *Server {
	s := &Server{router: mux.NewRouter(), source: source}
	s.routes()

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
