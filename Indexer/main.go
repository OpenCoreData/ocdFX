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
	"github.com/blevesearch/bleve"
	uuid "github.com/twinj/uuid"
)

// FileMetadata holds file metadata
type FileMetadata struct {
	MD5                   [16]byte
	UUID                  string
	filename              string
	LenContentNoStopWords int // will become content later
	CSDCOProjName         string
	CSDCOProjURI          string
}

func main() {
	flag.Parse()
	root := flag.Arg(0)

	// err := filepath.Walk(root, visit)
	// fmt.Printf("filepath.Walk() returned %v\n", err)

	size, err := dirSize(root)
	fmt.Printf("dirSize returned %d %v\n", size, err)
}

func dirSize(path string) (int64, error) {
	// open file system for triples

	// open a new Bleve index
	mapping := bleve.NewIndexMapping()
	// analyzer := mapping.Ad
	index, berr := bleve.New("csdcoFX.bleve", mapping)
	if berr != nil {
		fmt.Printf("Bleve error making index %v \n", berr)
	}

	// TODO Create the triple store set

	var size int64
	err := filepath.Walk(path, func(_ string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			size += f.Size()
			fileInfo := FileMetadata{}

			// Not sure why when doing a closure I need to rebuild the path name...
			FQP := fmt.Sprintf("%s/%s", path, f.Name())

			// md5
			data, _ := ioutil.ReadFile(FQP)
			fileInfo.MD5 = md5.Sum(data)

			// uuid
			u4 := uuid.NewV4()
			fileInfo.UUID = u4.String()

			// content via Tika
			url := "http://localhost:9998/tika"
			req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
			req.Header.Set("Accept", "text/plain")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
			}
			defer resp.Body.Close()

			// filter out some files that we don't want to index? dot files, what else?
			fmt.Println("Response Status:", resp.Status)
			body, _ := ioutil.ReadAll(resp.Body)
			cleanBody := stopwords.CleanString(string(body), "en", true)
			fileInfo.LenContentNoStopWords = utf8.RuneCountInString(cleanBody)
			dir, file := filepath.Split(FQP)

			berr = index.Index(fileInfo.UUID, fileInfo)
			if berr != nil {
				fmt.Printf("Bleve error indexing %v \n", berr)
			}

			// TODO Build the triples here and then append to the master set

			// // split so I can use a slice element in a lookup for metadata
			// fmt.Printf("%q\n", strings.Split(path, "/"))

			fmt.Printf("For path:\t%s \nFor dir:\t%s\nFor file:\t%s\nMD5:\t%v \nUUID:\t%s \nContentLen:\t%v \n\n", FQP, dir, file, fileInfo.MD5, fileInfo.UUID, fileInfo.LenContentNoStopWords)
		}
		return err
	})

	// TODO close out and serialize the triples to a file...

	return size, err
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
	dir, file := filepath.Split(path)

	// // split so I can use a slice element in a lookup for metadata
	// fmt.Printf("%q\n", strings.Split(path, "/"))

	fmt.Printf("For path:\t%s \nFor dir:\t%s\nFor file:\t%s\nMD5:\t%v \nUUID:\t%s \nContentLen:\t%v \n\n", path, dir, file, fileInfo.MD5, fileInfo.UUID, fileInfo.LenContentNoStopWords)

	//     fmt.Printf("is not dir: %s\n", path)
	// fmt.Printf("Found: %s  %d\n", f.Name(), f.Size())

	// indexJSON()  // pass to Bleve for indexing
	// generateTriples() // build triples and append to a flat file for later use in triple store

	return nil
}
