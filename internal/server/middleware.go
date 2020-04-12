package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) logRequest(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r)

		h(w, r)
	}
}

func (s *Server) logTileRequest(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		log.Println(vars["row"], vars["col"], vars["z"])

		h(w, r)
	}
}
