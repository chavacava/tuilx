package source

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ESAMapsProjection string

type ESAMaps struct {
	urlTemplate string
}

func NewESAMaps(layer string) *ESAMaps {
	var url = "https://tiles.esa.maps.eox.at/wmts/1.0.0/" + layer + "/default/WGS84/%d/%d/%d.jpg"

	return &ESAMaps{url}
}

func (ESAMaps) Name() string {
	return "esa-maps"
}

func (f ESAMaps) GetTile(ctx context.Context, col, row, zoom uint64) ([]byte, error) {
	url := fmt.Sprintf(f.urlTemplate, zoom, col, row)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
