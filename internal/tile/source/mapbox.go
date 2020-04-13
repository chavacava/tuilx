package source

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io/ioutil"
	"net/http"

	"github.com/chavacava/tuilx/internal/tile"
)

type MapBox struct {
	urlTemplate string
	name        string
}

func NewMapBox(layer, accessToken string) *MapBox {
	var url = "https://api.mapbox.com/styles/v1/mapbox/" + layer + "/tiles/%d/%d/%d/%d?access_token=" + accessToken

	return &MapBox{url, "mapbox/" + layer}
}

func (m MapBox) Name() string {
	return m.name
}

func (f MapBox) GetTile(ctx context.Context, props tile.Properties) (*tile.Tile, error) {
	url := fmt.Sprintf(f.urlTemplate, props.Size, props.Zoom, props.Col, props.Row)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[%s] error while resolving \"%s\", got status code: %v", f.Name(), url, resp.Status)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	dd, _, err := image.Decode(bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	return &tile.Tile{props, dd}, nil
}
