package middleware

import (
	"context"
	"fmt"
	"log"

	"github.com/chavacava/tuilx/internal/tile"
	"github.com/chavacava/tuilx/internal/tile/source"
)

// Cacher is a tile retriever that uses a cache
type Cacher struct {
	cache    Cache
	fallback source.Source
}

// Cache is the interface of cache implementations expected by the Cacher
type Cache interface {
	Get(key interface{}) (value interface{}, found bool)
	Set(key, value interface{})
}

// New yields a fresh new Cacher tile retriever
func NewCacher(cache Cache, fallback source.Source) Cacher {
	return Cacher{cache: cache, fallback: fallback}
}

func (Cacher) Name() string {
	return "cache"
}

func (c Cacher) GetTile(ctx context.Context, props tile.Properties) (*tile.Tile, error) {
	// get tile from cache
	cacheKey := fmt.Sprintf("%d/%d/%d", props.Zoom, props.Col, props.Row)
	data, found := c.cache.Get(cacheKey)
	if found {
		log.Printf("Returning tile from the cache")
		tile := data.(tile.Tile)
		return &tile, nil
	}

	// if fails, get it from the fallback retriever
	tile, err := c.fallback.GetTile(ctx, props)
	log.Printf("Returning tile from the cache fallback")
	// update cache (async?)
	go c.cache.Set(cacheKey, tile)

	// return the tile
	return tile, err
}
