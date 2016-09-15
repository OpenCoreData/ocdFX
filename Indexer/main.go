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
	"unicode/utf8"

	"github.com/bbalet/stopwords"
	uuid "github.com/twinj/uuid"
)

// FileMetadata holds file metadata
type FileMetadata struct {
	MD5                   [16]byte
	UUID                  string
	LenContentNoStopWords int
}

func main() {
	flag.Parse()
	root := flag.Arg(0)

	err := filepath.Walk(root, visit)
	fmt.Printf("filepath.Walk() returned %v\n", err)
}

func visit(path string, f os.FileInfo, err error) error {

	fileInfo := FileMetadata{}

	if f.IsDir() {
		dir := filepath.Base(path)
		fmt.Printf("Dir base path: %s \n", dir)
		return nil
	}

	// m5
	data, _ := ioutil.ReadFile(path)
	fileInfo.MD5 = md5.Sum(data)

	// uuid
	u4 := uuid.NewV4()
	fileInfo.UUID = u4.String()

	// content via Tika
	url := "http://localhost:9998/tika"
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
	// req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Accept", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// filter out some files that we don't want to index? dot files, what else?
	// filter out stop words and numbers?  No need to index these.
	fmt.Println("Response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	cleanBody := stopwords.CleanString(string(body), "en", true)
	fileInfo.LenContentNoStopWords = utf8.RuneCountInString(cleanBody)

	fmt.Printf("For file: %s \nMD5:\t%v \nUUID:\t%s \nContentLen:\t%v \n\n", path, fileInfo.MD5, fileInfo.UUID, fileInfo.LenContentNoStopWords)

	//  NEED the following SEVEN functions to be written...  all should be relatively easy...
	//  Call to triple store a simple HTTP call with a teplated SPARQL query
	// filemeta := GetFileMeta(path string, f os.FileInfo)  // Tika call  (do we need this?)
	// x contentText := GetFileContent()  // Tike call
	// projMeta := GetProjectMeta()  // SPARQL call

	// indexJSON()  // pass to Bleve for indexing
	// generateTriples() // build triples and append to a flat file for later use in triple store

	return nil
}
