#ocdFX
###An indexing and file exchange service for CSDCO files

####Intro
Initially for testing (and perhaps even for operation, we can use the docker file at
https://hub.docker.com/r/logicalspark/docker-tikaserver/ to run tika in server mode.  This container effectively runs: java -jar tika-server-x.x.jar -h 0.0.0.0


```
docker pull logicalspark/docker-tikaserver
```
We should be able to access this services via curl then for example.

```
curl -T ../OCD_METADATA.xls  http://localhost:9998/tika --header "Accept: text/plain"

curl -T testWORD.doc http://example.com:9998/meta
Returns the meta data as key/value pairs one per line. You can also have Tika return the results as JSON by adding the proper accept header:

curl -H "Accept: application/json" -T testWORD.doc http://example.com:9998/meta
[Update 2015-01-19] Previously the comment said that tika-server.jar is not available as download. Fixed that since it actually does exist as a binary download.


```

We can then build out our Go code to walk the files and build out the metadata 
that we need.  

####Indexer
For the indexer we will want to provide the ability to walk the directory structure and build out the necessary metadata we want to index on.  For us this will be something like the following sequence:

* Use Tika get file metadata.  While not the most useful for discovery, still worth getting and indexing
* Use Tika to get the file content text for supported file formats.  We will use this in the indexing as well.
* Use the directory structure to associate a file with its project.  We will then index all the project metadata with the file as well.  Place this in a "project" node so it can be searched on in the filter by:  project:XXXX
* Get the MD5 checksum of the file.  This will let us resolve if a file as changed or not and needs to be versioned and indexed.
* Assign a local in house UUID for this file.  Then associate this with its location.  This will be the primary key we use to dereference the file we want and obtain its bitstream.

####Questions
* A collection of files in a project will be called a "data collecton" or should be call this a "data set"?  Allow each file and or a zip collection of the files to be accessed?
* Should each file get a node type of something like gl:datastream or something to designate it a file?  Same for collections?


####Data Stores
This is just a collection of thoughts on this for now.  

* Should the generated JSON document in the Indexer section above be stored in MongoDB as well as indexed by Bleve?  Or should be just index it and not worry about keeping it?  The fact we have an MD5 would tell us if we need to rebuild the index or not.  What purpose is there to having the JSON as well?  Perhaps converting it to JSON-LD and allowing graph based relations between the result sets?  Yes, this is likely valuable and easy to do once in RDF.  So, the metadata structure should be:
	* converted to JSON and indexed in Bleve
	* converted to RDF and stored in the triple store
* For the quick FX aspect should be use BoltDB as a KV store to host the UUID to local file location information?  This would be fast, but if we accept hosting this in the triple store, wouldn't it be better to simply use that as the look up store?  For simplicity it is, so for now plan to use the triple store to resolve UUDI to local file location in ocdFX


###Refs
* JAXRS http://wiki.apache.org/tika/TikaJAXRS
* http://www.tutorialspoint.com/tika/tika_quick_guide.htm
* https://lucidworks.com/blog/2009/09/02/content-extraction-with-tika/
* Formats https://tika.apache.org/1.4/formats.html
* Tika app usage: https://tika.apache.org/1.11/gettingstarted.html



