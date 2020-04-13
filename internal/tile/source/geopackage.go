package source

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"image"
	"os"
	"time"

	"github.com/chavacava/tuilx/internal/tile"
	_ "github.com/mattn/go-sqlite3"
	"github.com/patrickmn/go-cache"
	"go.opencensus.io/trace"
)

type filePathBuilder func(ctx context.Context, props tile.Properties) string

// GeoPackage implements a tile source by retrieving tiles from geopackage
type GeoPackage struct {
	fpb         filePathBuilder
	queryPrefix string
	cache       *cache.Cache
}

// NewGeoPackage yields a fresh new GeoPackage tile source
func NewGeoPackage(fpb filePathBuilder, tableName string) *GeoPackage {
	c := cache.New(1*time.Minute, 5*time.Minute)
	c.OnEvicted(func(_ string, v interface{}) {
		conn := v.(*sql.DB)
		_ = conn.Close()
	})
	return &GeoPackage{
		fpb:         fpb,
		queryPrefix: "SELECT tile_data FROM " + tableName + " WHERE zoom_level=",
		cache:       c,
	}
}

func (GeoPackage) Name() string {
	return "geopackage"
}

func (g GeoPackage) GetTile(ctx context.Context, props tile.Properties) (*tile.Tile, error) {
	gpkgFile := g.fpb(ctx, props)
	if _, err := os.Stat(gpkgFile); err != nil {
		return nil, fmt.Errorf("unable to access file %s: %v", gpkgFile, err)
	}

	var dbHandler *sql.DB
	value, ok := g.cache.Get(gpkgFile)
	switch ok {
	case true:
		dbHandler = value.(*sql.DB)
	case false:
		var err error
		_, span := trace.StartSpan(ctx, "db connect")
		dbHandler, err = sql.Open("sqlite3", gpkgFile+"?immutable=true")
		span.End()
		if err != nil {
			return nil, err
		}
	}

	_, span := trace.StartSpan(ctx, "db query")
	qs := g.queryPrefix + fmt.Sprintf("%d AND tile_column=%d AND tile_row=%d LIMIT 1", props.Zoom, props.Col, props.Row)
	rows, err := dbHandler.Query(qs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tileData := []byte{}
	if rows.Next() {
		err := rows.Scan(&tileData)
		if err != nil {
			return nil, err
		}
	}
	span.End()
	_, span = trace.StartSpan(ctx, "image decode")
	dd, _, err := image.Decode(bytes.NewBuffer(tileData))
	span.End()
	if err != nil {
		return nil, err
	}

	// TODO set actual tile properties
	result := tile.Tile{Props: props, Data: dd}
	return &result, nil
}
