# lazyosm

# What is it? 

Lazyosm is something I've been half attempting to do for a while, it attempts to create a relational model of all the file blocks and the underlying osm pbf file structure to properly utilize assembling features as quickly as possible. This differs drastically from other go implementations of osm data that purely go for memory through of primitive features which doesn't do you a whole lot of good if your node for a way is 100 million primitive features from one another. 

I think what I'm trying to say is this methodology isn't probably the way its intended to be used, throwing all your nodes in a k,v databases and swapping each data structure out to build features is silly (your deserializing nodes just to be written to a file again) but people do it. More importanty the osm file structure I think gives you a few clues on how to do it in a more nuanced manner, but I could be wrong. 

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

Relations I'm kicking around but I think I know what I'm going to do. 


# Mapping functionality

Obviously the goal of this is to abstract to something like importosm3 with a mapping structure and all of that. This shouldn't be hard again just something I have to think about implementing correctly. 

# Performance 

Don't ask IDK. From just eyeballing it against other things I've seen it seems like its really fast and this is completely unoptimized, currently just trying to get things working. 
