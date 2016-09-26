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
	rdf "github.com/knakk/rdf"
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
	tr := []rdf.Triple{}

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
			newsub, _ := rdf.NewIRI(fmt.Sprintf("http://opencoredata/id/resource/csdco/datafile/%s", fileInfo.UUID)) // Sprintf a correct URI here
			newpred0, _ := rdf.NewIRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#type")
			newobj0, _ := rdf.NewIRI("http://opencoredata.org/id/voc/csdco/v1/Datafile")
			newtriple0 := rdf.Triple{Subj: newsub, Pred: newpred0, Obj: newobj0}

			newpred1, _ := rdf.NewIRI("http://opencoredata.org/id/voc/csdco/v1/fileuuid")
			newobj1, _ := rdf.NewLiteral(fileInfo.UUID)
			newtriple1 := rdf.Triple{Subj: newsub, Pred: newpred1, Obj: newobj1}

			newpred2, _ := rdf.NewIRI("http://opencoredata.org/id/voc/csdco/v1/size")
			newobj2, _ := rdf.NewLiteral(fileInfo.LenContentNoStopWords)
			newtriple2 := rdf.Triple{Subj: newsub, Pred: newpred2, Obj: newobj2}

			newpred3, _ := rdf.NewIRI("http://opencoredata.org/id/voc/csdco/v1/name")
			newobj3, _ := rdf.NewLiteral(file)
			newtriple3 := rdf.Triple{Subj: newsub, Pred: newpred3, Obj: newobj3}

			newpred4, _ := rdf.NewIRI("http://opencoredata.org/id/voc/csdco/v1/Project")
			newobj4, _ := rdf.NewIRI("project URI here")
			newtriple4 := rdf.Triple{Subj: newsub, Pred: newpred4, Obj: newobj4}

			// will any every need to be skipped to add triples?
			tr = append(tr, newtriple0)
			tr = append(tr, newtriple1)
			tr = append(tr, newtriple2)
			tr = append(tr, newtriple3)
			tr = append(tr, newtriple4)
			// // split so I can use a slice element in a lookup for metadata
			// fmt.Printf("%q\n", strings.Split(path, "/"))

			fmt.Printf("For path:\t%s \nFor dir:\t%s\nFor file:\t%s\nMD5:\t%v \nUUID:\t%s \nContentLen:\t%v \n\n", FQP, dir, file, fileInfo.MD5, fileInfo.UUID, fileInfo.LenContentNoStopWords)
		}
		return err
	})

	// TODO close out and serialize the triples to a file...
	fmt.Println(tr)

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
