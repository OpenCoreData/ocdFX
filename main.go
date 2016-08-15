package main

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bbalet/stopwords"
	uuid "github.com/twinj/uuid"
)

func visit(path string, f os.FileInfo, err error) error {

	if f.IsDir() {
		dir := filepath.Base(path)
		fmt.Printf("Dir base path: %s \n", dir)
		return nil
	}

	fmt.Printf("Visited file: %s\n", path)

	//m5
	data, _ := ioutil.ReadFile(path)
	fmt.Printf("MD5:  %x", md5.Sum(data))

	//uuid
	u4 := uuid.NewV4()
	fmt.Println(u4)
	fmt.Printf("UUID: version %d variant %x: %s\n", u4.Version(), u4.Variant(), u4)

	//content POST
	// req, err := sling.New().Post("http://upload.com/gophers")
	url := "http://localhost:9998/tika"

	// filter out some files that we don't want to index? dot files, what else?
	// filter out stop words and numbers?  No need to index these.
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
	// req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Accept", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	cleanBody := stopwords.CleanString(string(body), "en", true)
	fmt.Println("response Body:", cleanBody)

	//  NEED the following SEVEN functions to be written...  all should be relatively easy...
	//  Calls to Tike will be REST client calls.
	//  Call to triple store a simple HTTP call with a teplated SPARQL query
	// filemeta := GetFileMeta(path string, f os.FileInfo)  // Tika call  (do we need this?)
	// x contentText := GetFileContent()  // Tike call
	// projMeta := GetProjectMeta()  // SPARQL call
	// x md5value := GetMD5Value()    // function call
	// x fileUUID := GenerateUUID()   // function call

	// data := []byte("These pretzels are making me thirsty.")
	//
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
