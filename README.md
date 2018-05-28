# lazyosm - Lazily creates end user geojson features from osm datasets 

This project could be much more mature than it currently is, I think all the stuff it does is super cool, but is a little to verbose to explain in a read me. Basically lazyosm reads each file block containing 8000 primitive osm features lazily maps those to id ranges and enables one to filter, build, and write out end user osm features as geojson much much faster than any methodology I've seen. 

So instead of say reading osm primitive features as fast as possible and immediately writing out to some key value store. Billions of primitive osm features could be inbetween you and more complex features. Instead we map all the ids pertaining to densenodes, ways, and relations in each fileblock than deduce which node fileblocks correspond to a single way fileblock from there we can read in a lean nodemap structure to build out the way within the single way fileblock. 

Beyond just that you can start implementing things like complex mapping structures to the string tables of each file block.
Lazy computation enables you to do some pretty crazy stuff you normally wouldn't even be able to consider. 

#### Maturity

As far as feature building and writing out to a geobuf file (geojson) this project is pretty far along, the only issues that may arise could be in tags indicating what type of feature they actually are, and potentially some multipolygon edge cases I haven't sniffed out yet. (although it actually does pretty good imo) 

#### Why Haven't I Invested More Time In This 

Well first off I work in a Civil Engineering not software so I'm sort of relegated to weekends however, the real reason I've sort of stopped development on this project is I see the convergance of a few different file formats and geographical representations that I want to build a stack around. (geobuf,vector-tiles,geojson-vt features) Generally I kind of just build something but I feel like a more pragmatic approach to something like feature filtering could save me a lot of trouble in the future. 


