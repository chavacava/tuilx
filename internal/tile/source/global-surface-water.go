package source

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io/ioutil"
	"net/http"

	"github.com/chavacava/lab-tileserver/internal/tile"
)

type WSW struct {
	urlTemplate string
}

func NewWSW() *WSW {
	var url = "https://storage.googleapis.com/global-surface-water/tiles2018/transitions/%d/%d/%d.png"
	return &WSW{url}
}

func (WSW) Name() string {
	return "global-surface-water"
}

func (f WSW) GetTile(ctx context.Context, props tile.Properties) (*tile.Tile, error) {
	url := fmt.Sprintf(f.urlTemplate, props.Zoom, props.Col, props.Row)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[%s] error while resolving \"%s\", got status code: %v", f.Name(), url, resp.Status)
	}
	defer resp.Body.Close()

	tileData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	dd, _, err := image.Decode(bytes.NewBuffer(tileData))
	if err != nil {
		return nil, err
	}
	// TODO set actual tile properties
	result := tile.Tile{props, dd}
	return &result, nil
}
