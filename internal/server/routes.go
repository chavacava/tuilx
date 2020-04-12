package server

import (
	"net/http/pprof"
)

const (
	FieldZoom   = "z"
	FieldCol    = "col"
	FieldRow    = "row"
	FieldSize   = "size"
	FieldFormat = "format"
)

func (s *Server) routes() {
	s.router.HandleFunc("/tiles/{"+FieldSize+"}/{"+FieldZoom+"}/{"+FieldCol+"}/{"+FieldRow+"}.{"+FieldFormat+"}", s.logTileRequest(s.handleGetTile()))
}

func (s *Server) AttachProfiler() {
	s.router.HandleFunc("/debug/pprof/", pprof.Index)
	s.router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	s.router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	// Manually add support for paths linked to by index page at /debug/pprof/
	s.router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	s.router.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	s.router.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	s.router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	s.router.Handle("/debug/pprof/flamegraph", pprof.Handler("flamegraph"))
	s.router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	s.router.Handle("/debug/pprof/block", pprof.Handler("block"))
}
