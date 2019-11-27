# lazyosm - Lazily creates end user geojson features from osm datasets 
[![GoDoc](https://img.shields.io/badge/api-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/murphy214/lazyosm)

This project could be much more mature than it currently is, I think all the stuff it does is super cool, but is a little to verbose to explain in a read me. Basically lazyosm reads each file block containing 8000 primitive osm features lazily maps those to id ranges and enables one to filter, build, and write out end user osm features as geojson much much faster than any methodology I've seen. 

So instead of say reading osm primitive features as fast as possible and immediately writing out to some key value store. Billions of primitive osm features could be inbetween you and more complex features. Instead we map all the ids pertaining to densenodes, ways, and relations in each fileblock than deduce which node fileblocks correspond to a single way fileblock from there we can read in a lean nodemap structure to build out the way within the single way fileblock. 

Beyond just that you can start implementing things like complex mapping structures to the string tables of each file block.
Lazy computation enables you to do some pretty crazy stuff you normally wouldn't even be able to consider. 

#### Maturity

As far as feature building and writing out to a geobuf file (geojson) this project is pretty far along, the only issues that may arise could be in tags indicating what type of feature they actually are, and potentially some multipolygon edge cases I haven't sniffed out yet. (although it actually does pretty good imo) 

### Install 

```
go get -u github.com/murphy214/lazyosm
```

### CLI - Usage 

For a minimal example I implemented a cli interface to dump raw geojson features to a file from a pbf geofabrik file. Download the vermont pbf [here](http://download.geofabrik.de/north-america/us/vermont-latest.osm.pbf). Then navigate to where that file is located in your terminal.

```
lazyosm make -f vermont-latest.osm.pbf -o vermont-out.geojson
```

Will create the applicable geojson file. 




