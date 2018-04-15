# lazyosm

# What is it? 

Lazyosm is something I've been half attempting to do for a while, it attempts to create a relational model of all the file blocks and the underlying osm pbf file structure to properly utilize assembling features as quickly as possible. This differs drastically from other go implementations of osm data that purely go for memory through of primitive features which doesn't do you a whole lot of good if your node for a way is 100 million primitive features from one another. 

## What it does currently

Currently its uses are pretty limited, the only output that exists for pure features is a geobuf output format in which features are iteratively written, as it sits this is an extremely bare bones api and doesn't have end points for things like mapping yet. This library has uses that go beyond just pushing a mapping into an sql though, a flat map that can be utilized for each block and get a specific kind of feature can be implemented as well. 

# Caveats

Corner cases exist in multipolygons, and I have no idea what the hell I'm doing with area / specific tags from ways other than area = 'yes' to many corner cases currently, and the osm documentation for what tags dictate what are sort of esoteric. 

# Implementation Notes

I've hacked on this quite a bit so here are some scattered thoughts about what does what how and why. 

Without going deep into osm file structure (I'm not super familiar with it anyway) osm data is effectively a set of file block each containing 8000 primitive features either being densenodes,nodes,ways, or relations. These blocks always have the nodes first, the dense nodes second etc            

### The IdMap

The IdMap is a data structure made from an upper bound and lower bound of ids its job is for a given node id relate it back to the file position accordingly. Currently I'm only dealing with files that have several thousand file blocks but in the future this could get expensive to iterate through an entire list everytime. The IdMap simply uses a math.Floor() implementation of iteger indexs against several log values to derive indexs a little more quickly.

**TL:DR a data structure to go from an id to a file block it exists exists in**

### The NodeID map

This data structure is pretty complex. It is a stateful node map of the last x blocks used with the assurance that x blocks can accomedate all values in a way file block. A nodemap has the data structure map[int][int][]float64 where the first map is for the file block it resides in and the second int is for id itself. The first map exists so we have a hinge to delete that nodemap in bulk. This nodemaps are also lazily read so literally the only part were reading from a densenode set is the ids lats and longs exactly what we need no unnecessary allocations. 

**TL:DR Here we constantly reading into memory and deleting out of memory different node maps. We gather which nodes need to come out of memory from a priority queue like structure when the map gets above the allowed limit. This process is usually done in chunks of way blocks or a single way block currently.**

Need to get a few pesky buts though. 


# Relations 

DONE - Corners cases exist and needs to be optimized pretty slow. 

# Mapping functionality

NOT DONE - relatively easy other than the code being in a few different places

# Performance 

**I can pretty confidently say that for the use case of limited size memory and disk space, that this should be pretty easily the fastest implementation for traversing pbf files and creating features at least in go. Especially when I implement string table scan against blocks lazily, it should easily be the best for broad searches across large pbf files for specific tagsets.**

The largest chunk of time generally speaking for a read intake to geobuf is generally the way node traversal as this has the largest amount of i/o between different node blocks. To remedy this I implemented an algorithm that calculates the shortest path for traversing a set way nodes given the node id blocks that exist within them. This basically simplifies or translates to a traveling sales man like algorithm in which we are optimizing for block i/o. (i.e. read / write / gc) In order to speed up the shortest path alg. assumptions were made and shortcuts were taken because if not the shortest path alg. could end up eating a substantial amount of computing time as its a stateful algorithm and can't be easily parrelized. (i.e. the last way added is needed to see which is optimal to come next) 

# To-Dos 

* Its really shitty I have so many dependencies for this project, but many of them are pretty essential geobuf for faster writes geojson as its basically my end structure, and pbf to read the raw pbf values lazily are all pretty essential to how this library is implemented. The only one maybe worth dropping is mercantile as its not really used in a meanful way.

* Clean up code implementations are scattered all of over the place. 
