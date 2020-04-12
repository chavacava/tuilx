package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/chavacava/lab-tileserver/internal/tile"
	"github.com/gorilla/mux"
	"go.opencensus.io/trace"
)

func (s *Server) handleRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("In the root!"))
	}
}

func (s *Server) handleGetTile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := trace.StartSpan(context.Background(), "getTile handler")
		defer span.End()

		tileProps, err := extractTileProperties(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		tile, err := s.source.GetTile(ctx, tileProps)
		if err != nil {
			log.Println(err.Error())
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("Internal Error while serving %V+", r))
			return
		}

		encoded, err := tile.TranscodeData(tileProps.Encoding)
		if err != nil {
			log.Println(err.Error())
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("Internal Error while serving %v+", r))
			return
		}

		w.Write(encoded)
	}
}

func extractTileProperties(r *http.Request) (props tile.Properties, err error) {
	vars := mux.Vars(r)

	cs, exists := vars[FieldCol]
	if !exists {
		err = errors.New("tile column not specified in the request")
		return
	}
	rs, exists := vars[FieldRow]
	if !exists {
		err = errors.New("tile row not specified in the request")
		return
	}
	zs, exists := vars[FieldZoom]
	if !exists {
		err = errors.New("tile zoom not specified in the request")
		return
	}
	f, exists := vars[FieldFormat]
	if !exists {
		err = errors.New("tile format not specified in the request")
		return
	}
	ss, exists := vars[FieldSize]
	if !exists {
		err = errors.New("tile size not specified in the request")
		return
	}

	col, err := strconv.ParseUint(cs, 0, 64)
	if err != nil {
		err = errors.New("tile column must be uint, got '" + cs + "'")
		return
	}
	row, err := strconv.ParseUint(rs, 0, 64)
	if err != nil {
		err = errors.New("tile row must be uint, got '" + rs + "'")
		return
	}
	zoom, err := strconv.ParseUint(zs, 0, 64)
	if err != nil {
		err = errors.New("tile zoom must be uint, got '" + zs + "'")
		return
	}
	size, err := strconv.ParseUint(ss, 0, 16)
	if err != nil {
		err = errors.New("tile size must be uint16, got '" + zs + "'")
		return
	}

	switch f {
	case "png":
		props.Encoding = tile.PNG
	case "jpg", "jpeg":
		props.Encoding = tile.JPEG
	default:
		err = errors.New("unhandled tile format '" + f + "'")
		return
	}
	props.Col = col
	props.Row = row
	props.Zoom = zoom
	props.Size = uint16(size)

	return props, nil
}

func respondError(w http.ResponseWriter, errorCode int, msg string) {
	w.WriteHeader(errorCode)
	w.Write([]byte(msg))
}
