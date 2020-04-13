# TuilX

An extensible and composable WMTS server.

TuilX allows you to easily setup a WMTS server by creating and/or combining _tile sources_.

TuilX provides some basic sources like:

| Source | Description |
|--------|-------------|
| Geopackage | Reads WMTS tiles from geopackage files|
| Mapbox | Retrieves WMTS tiles from MapBox |
| ESAMaps | Retrieves WMTS tiles from the European Space Agency services |

There are also _utility sources_ (aka _middleware_) like:
| Source | Description |
|--------|-------------|
| Cacher | Adds a cache to a source |
| Masker | Applies masking to tiles |
| Resizer | Resizes tiles |
| Fallback | Retrieves a tile from a fallback source if the primary source fails |
| Tracer | Trace (opencensus) the retrieval of a tile from a source|
| Switcher | Switches among sources depending on conditions |

The list of sources is potentially infinite, you can develop your own: any `struct` satisfying the following interface is a _tile source_:

```go
type Source interface {
	GetTile(ctx context.Context, props tile.Properties) (*tile.Tile, error)
	Name() string
}
```
To setup your server you just need to provide a source:
```go
// Create the web server
server := server.New(mySource)
```
where `mySource` is any of the basic sources or a combination of them. For example:

```go
primary := source.NewESAMaps("s2cloudless-2018", esa-auth-token)
fb := source.NewMapBox("satellite-v9", mb-auth-token)
mySource := middleware.NewFallback(primary, fb)
```

will try to serve the requested tile from the Sentinel-2 ESA layer, and if retrieval fails, it will serve the tile from MapBox satellite layer.

If then you are interested in tracing the tile retrieval:

```go
primary := source.NewESAMaps("s2cloudless-2018", esa-auth-token)
fb := source.NewMapBox("satellite-v9", mb-auth-token)
mySource := middleware.Trace(middleware.NewFallback(primary, fb))
```

Then you may decide to add a cache to MapBox source:

```go
primary := source.NewESAMaps("s2cloudless-2018", esa-auth-token)
fb := middleware.NewCacher(myRedis, source.NewMapBox("satellite-v9", mb-auth-token))
mySource := middleware.Trace(middleware.NewFallback(primary, fb))
```

Take a look at `cmd/tuilx-svr.go` for a complete example.
