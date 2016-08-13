package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func visit(path string, f os.FileInfo, err error) error {

	if f.IsDir() {
		dir := filepath.Base(path)
		fmt.Printf("Dir base path: %s \n", dir)
		return nil
	}

	fmt.Printf("Visited file: %s\n", path)

	//  NEED the following SEVEN functions to be written...  all should be relatively easy...
	//  Calls to Tike will be REST client calls.
	//  Call to triple store a simple HTTP call with a teplated SPARQL query
	// filemeta := GetFileMeta(path string, f os.FileInfo)  // Tika call
	// contentText := GetFileContent()  // Tike call
	// projMeta := GetProjectMeta()  // SPARQL call
	// md5value := GetMD5Value()    // function call
	// fileUUID := GenerateUUID()   // function call

	// All the above should be in a struct..  from that JSON and RDF can be built.

	// indexJSON()  // pass to Bleve for indexing
	// generateTriples() // build triples and append to a flat file for later use in triple store

	return nil
}

func main() {
	flag.Parse()
	root := flag.Arg(0)
	err := filepath.Walk(root, visit)
	fmt.Printf("filepath.Walk() returned %v\n", err)
}
