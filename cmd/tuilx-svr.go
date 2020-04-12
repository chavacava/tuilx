package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"contrib.go.opencensus.io/exporter/zipkin"
	"github.com/chavacava/lab-tileserver/internal/server"
	"github.com/chavacava/lab-tileserver/internal/tile"
	"github.com/chavacava/lab-tileserver/internal/tile/source"
	"github.com/chavacava/lab-tileserver/internal/tile/source/middleware"
	ozipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/trace"
)

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		log.Fatalf("%s\n", err)
	}
}

func run(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		port             string
		at               string
		gpkgPathTemplate string
		gpkgTable        string
		zipkinServer     string
	)
	flags.StringVar(&port, "p", ":8080", "service port")
	flags.StringVar(&at, "mapbox-at", "", "mapbox authorization token")
	flags.StringVar(&gpkgPathTemplate, "gpkg-path", "UNDEFINED gpkg-path", "path template for geopackage files (zoom,col,row)")
	flags.StringVar(&gpkgTable, "gpkg-table", "tiles", "table name where tiles are stored")
	flags.StringVar(&zipkinServer, "trace-zkserver", "localhost:9411", "Zipkin server host and port where to send traces")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if zipkinServer != "" {
		setupTracing(zipkinServer, port)
	}

	// You can build your own server from here

	// The following is just an example server with three main sources:
	// 1. a MainMap that is stored in geopackage files
	// 2. a Bluemarble-like layer retrieved form MapBox
	// 3. a WaterMask layer retrieved from another WMTS service
	//
	// When the requested tile is under the zoom 8, we will serve a Bluemarble tile
	// For zooms 8 and higher we will serve a MainMap tile, water pixels will be replaced by pixels from the Bluemarble
	// this replacement is only performed in zooms upt to 13 (the max zoom of the WaterMask)

	fpb := func(ctx context.Context, props tile.Properties) string {
		const lowestZoom = 8
		div := uint64(1 << (props.Zoom - lowestZoom))
		if div < 1 {
			div = 1
		}
		r := props.Row / div
		c := props.Col / div
		return fmt.Sprintf(gpkgPathTemplate, lowestZoom, c, r)
	}

	// MainMap tiles are retrieved from geopackages
	// this tile source will be traced
	mainMap := middleware.Trace(source.NewGeoPackage(fpb, gpkgTable))
	// Lower res tiles are retrieved from mapbox
	// this tile source will be traced
	blueMarble := middleware.Trace(source.NewMapBox("satellite-v9", at))

	// Watermask raster tiles are retrived from a globals surface water provider
	// tile source will be traced
	waterMask := middleware.Trace(source.NewWSW())

	// MainMap tiles will be masked
	// this tile source will be traced
	maskedMainMap := middleware.Trace(middleware.NewMasker(
		mainMap,
		waterMask,
		blueMarble))

	// The server will switch between tile sources depending on the requested zoom level
	//  Criteria on requested zoom level
	zoomSwitcher := func(_ context.Context, props tile.Properties) source.Source {
		switch z := props.Zoom; {
		case z > 13:
			return mainMap
		case z > 7:
			return maskedMainMap
		default:
			return blueMarble
		}
	}
	//  Configure the switch
	zoomConditioned := middleware.Trace(middleware.NewSwitcher("zoom-swt", zoomSwitcher))

	// Create the web server
	server := server.New(zoomConditioned)
	server.AttachProfiler()

	return http.ListenAndServe(port, server)
}

// setupTracing setups the tracing infrastructure
func setupTracing(host string, thisServicePort string) {
	localEndpoint, err := ozipkin.NewEndpoint("tileserver", "localhost"+thisServicePort)
	if err != nil {
		log.Fatalf("Failed to create the local zipkin endpoint: %v", err)
	}

	reporter := zipkinHTTP.NewReporter("http://" + host + "/api/v2/spans")
	ze := zipkin.NewExporter(reporter, localEndpoint)
	trace.RegisterExporter(ze)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
}
