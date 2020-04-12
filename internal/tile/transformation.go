package tile

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"

	"github.com/pkg/errors"
)

const defaultJPEGQuality = 100

func (t Tile) TranscodeData(target EncodingType) (result []byte, err error) {
	buf := new(bytes.Buffer)
	switch target {
	case PNG:
		if err := png.Encode(buf, t.Data); err != nil {
			return nil, errors.Wrapf(err, "unable to transcode %v as as %s", t.Data.ColorModel, PNG)
		}
	case JPEG:
		// use default JPEG options
		if err := jpeg.Encode(buf, t.Data, nil); err != nil {
			return nil, errors.Wrapf(err, "unable to transcode %s as %s", t.Data.ColorModel, JPEG)
		}
	default:
		return nil, fmt.Errorf("don't know how to transcode into %s", target)
	}

	return buf.Bytes(), nil
}
