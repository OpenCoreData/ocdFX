package main

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bbalet/stopwords"
	"github.com/blevesearch/bleve"
	rdf "github.com/knakk/rdf"
	sparql "github.com/knakk/sparql"
	uuid "github.com/twinj/uuid"
)

const queries = `
# Comments are ignored, except those tagging a query.

# tag: test1
SELECT DISTINCT *
WHERE
{
  ?uri rdf:type <http://opencoredata.org/id/voc/csdco/v1/CSDCOProject> .
  ?uri <http://opencoredata.org/id/voc/csdco/v1/project> "{{.}}" .
  ?uri ?p ?o .
}

# tag: urionly
SELECT DISTINCT ?uri
WHERE
{
  ?uri rdf:type <http://opencoredata.org/id/voc/csdco/v1/CSDCOProject> .
  ?uri <http://opencoredata.org/id/voc/csdco/v1/project> "{{.}}" .
  ?uri ?p ?o .
}

`

// FileMetadata holds file metadata
type FileMetadata struct {
	MD5                [16]byte
	UUID               string
	Filename           string
	Filetype           string
	ContentNoStopWords string // will become content later
	CSDCOProjName      string
	CSDCOProjURI       string
}

func main() {
	flag.Parse()
	root := flag.Arg(0)

	// err := filepath.Walk(root, visit)
	// fmt.Printf("filepath.Walk() returned %v\n", err)

	size, err := dirSize(root)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("dirSize returned %d %v\n", size, err)
}

func dirSize(path string) (int64, error) {
	// open file system for triples

	// open a new Bleve index
	mapping := bleve.NewIndexMapping()
	// analyzer := mapping.Ad
	index, berr := bleve.New("csdcoFX.bleve", mapping)
	berr = index.Close()
	if berr != nil {
		fmt.Printf("Bleve error making index %v \n", berr)
	}

	//Create the triple store set and size var
	// tr := []rdf.Triple{}
	var size int64

	err := filepath.Walk(path, func(fp string, f os.FileInfo, err error) error {
		if !f.IsDir() {

			// for fun see how much we index...  this is the file length in bytes (int64)
			// 1MB = 1000000 bytes
			size += f.Size() // use this to escape Tika indexing on files that are "too" big

			// set up our predicate value..  if we end up not setting a predicate it's
			// due to the fact we didn't match a file value and so we need to not index this
			// particular file
			var predicate string

			pathElements := strings.Split(fp, "/")
			projectIDSet := strings.Split(pathElements[5], " ")
			projectID := projectIDSet[0]

			// TODO..   work up dotfile escape element
			dotfile, err := filepath.Match(strings.ToLower(".*"), strings.ToLower(f.Name()))
			if dotfile {
				fmt.Println("Getting out..  found a dot file.")
				return nil // get out now..  we are a file with a . at the start
			}
			if err != nil {
				fmt.Println(err) // malformed pattern
				return err       // this is fatal.
			}

			// TODO test getting file stats
			if ageInYears(fp) < 2.0 {
				fmt.Println("Getting out..  found a file < 2 years old.")
				fmt.Println(fp)
				return nil // get out now..  we are a too young
			}

			// TODO  check a whitelist
			if !inApprovedList(projectID) {
				fmt.Println("Getting out..  found a file not in white list.")
				fmt.Println(fp)
				return nil // get out now...  we are not wanted
			}

			if caseInsenstiveContains(fp, "/") {

				// Looking for [ProjectID]-metadata
				// We really don't know the PID..  need to try and pull that from the path....
				matched, err := filepath.Match(strings.ToLower("*-metadata*"), strings.ToLower(f.Name()))
				if matched {
					predicate = "http://opencoredata.org/id/voc/csdco/v1/metadata"
				}

				// look for Dtube lable name...
				if !matched {
					matched, err = filepath.Match(strings.ToLower("*metadata format Dtube Label_*"), strings.ToLower(f.Name())) // worry about case issue
					if matched {
						predicate = "http://opencoredata.org/id/voc/csdco/v1/dtubeMetadata"
					}
				}

				// subsample metadata information
				if !matched {
					matched, err = filepath.Match(strings.ToLower("*SRF*"), strings.ToLower(f.Name()))
					if matched {
						predicate = "http://opencoredata.org/id/voc/csdco/v1/srf"
					}
				}

				// Corelyzer session file
				if !matched {
					matched, err = filepath.Match(strings.ToLower("*.cml"), strings.ToLower(f.Name()))
					if matched {
						predicate = "http://opencoredata.org/id/voc/csdco/v1/cml"
					}
				}

				// Corelyzer archive file
				if !matched {
					matched, err = filepath.Match(strings.ToLower("*.car"), strings.ToLower(f.Name()))
					if matched {
						predicate = "" // for now ignore .car files.  they are too to index and are archives big as are .mov
						// predicate = "http://opencoredata.org/id/voc/csdco/v1/car"
					}
				}
				if err != nil {
					fmt.Println(err) // malformed pattern
					return err       // this is fatal.
				}
				if matched {
					fmt.Printf("%s : %s : %s\n", projectID, predicate, fp)
				}
			}

			if caseInsenstiveContains(fp, "Images/") {
				matched, err := filepath.Match(strings.ToLower("*.jpg"), strings.ToLower(f.Name()))

				if !matched {
					matched, err = filepath.Match(strings.ToLower("*.jpeg"), strings.ToLower(f.Name()))
				}
				if !matched {
					matched, err = filepath.Match(strings.ToLower("*.tif"), strings.ToLower(f.Name()))
				}
				if !matched {
					matched, err = filepath.Match(strings.ToLower("*.tiff"), strings.ToLower(f.Name()))
				}
				if !matched {
					matched, err = filepath.Match(strings.ToLower("*.bmp"), strings.ToLower(f.Name()))
				}

				if err != nil {
					fmt.Println(err) // malformed pattern
					return err       // this is fatal.
				}
				if matched {
					predicate = "http://opencoredata.org/id/voc/csdco/v1/image"
					fmt.Printf("%s : %s : %s\n", projectID, predicate, fp)
				}
			}

			if caseInsenstiveContains(fp, "Images/rgb") {
				matched, err := filepath.Match(strings.ToLower("*.csv"), strings.ToLower(f.Name()))
				if err != nil {
					fmt.Println(err) // malformed pattern
					return err       // this is fatal.
				}
				if matched {
					predicate = "http://opencoredata.org/id/voc/csdco/v1/rgbData"
					fmt.Printf("%s : %s : %s\n", projectID, predicate, fp)
				}
			}

			if caseInsenstiveContains(fp, "Geotek Data/whole-core data") {

				// black list this extensions in here: .raw .dat .out and .cal
				matched, err := filepath.Match(strings.ToLower("*.raw"), strings.ToLower(f.Name()))
				if !matched {
					matched, err = filepath.Match(strings.ToLower("*.dat"), strings.ToLower(f.Name()))
				}
				if !matched {
					matched, err = filepath.Match(strings.ToLower("*.out"), strings.ToLower(f.Name()))
				}
				if !matched {
					matched, err = filepath.Match(strings.ToLower("*.cal"), strings.ToLower(f.Name()))
				}

				if err != nil {
					fmt.Println(err) // malformed pattern
					return err       // this is fatal.
				}
				if matched {
					return nil // We return nil on match here since we got a postive from the Black list above
				}

				// if we don't drop out in the above black list, check for our white list pattern
				matched, err = filepath.Match(strings.ToLower("*_MSCL*"), strings.ToLower(f.Name()))
				if matched {
					// now check for correct extensions
					matched, err = filepath.Match(strings.ToLower("*.xls"), strings.ToLower(f.Name()))
					if !matched {
						matched, err = filepath.Match(strings.ToLower("*.xlsx"), strings.ToLower(f.Name()))
					}
					if !matched {
						return nil // done with this test loop..
					}
					if matched {
						predicate = "http://opencoredata.org/id/voc/csdco/v1/wholeCoreData"
						fmt.Printf("%s : %s : %s\n", projectID, predicate, fp)
					}
					if err != nil {
						fmt.Println(err) // malformed pattern
						return err       // this is fatal.
					}
				}
			}

			if caseInsenstiveContains(fp, "Geotek Data/high-resolution MS data") {
				matched, err := filepath.Match(strings.ToLower("*_HRMS*"), strings.ToLower(f.Name()))

				if !matched {
					matched, err = filepath.Match(strings.ToLower("*_XYZ*"), strings.ToLower(f.Name()))
				}

				if err != nil {
					fmt.Println(err) // malformed pattern
					return err       // this is fatal.
				}
				if matched {
					// now check for correct extensions  //TODO  ask if I should add .csv here as well..

					matched, err = filepath.Match(strings.ToLower("*.xls"), strings.ToLower(f.Name()))
					if !matched {
						matched, err = filepath.Match(strings.ToLower("*.xlsx"), strings.ToLower(f.Name()))
					}
					if !matched {
						return nil // done with this test loop..
					}
					if matched {
						predicate = "http://opencoredata.org/id/voc/csdco/v1/geotekHighResMSdata"
						fmt.Printf("%s : %s : %s\n", projectID, predicate, fp)
					}
					if err != nil {
						fmt.Println(err) // malformed pattern
						return err       // this is fatal.
					}
				}

			}

			// Walk all subdirectories?
			if caseInsenstiveContains(fp, "ICD/") {
				matched, err := filepath.Match("ICD sheet.pdf", strings.ToLower(f.Name()))
				if matched {
					return nil // we matched above so get out now...
				}

				matched, err = filepath.Match(strings.ToLower("*.pdf"), strings.ToLower(f.Name()))

				if err != nil {
					fmt.Println(err) // malformed pattern
					return err       // this is fatal.
				}
				if matched {
					predicate = "http://opencoredata.org/id/voc/csdco/v1/icdFiles"
					fmt.Printf("%s : %s : %s\n", projectID, predicate, fp)
				}
			}

			// start a if conditional here based on if we have a predicate value or not.  If
			// we do, we matched above...
			if predicate != "" {

				// for incremental file saving move tr inside the loop (and the write function at then of this scope)
				tr := []rdf.Triple{}

				// our struct for information
				fileInfo := FileMetadata{}

				// Not sure why when doing a closure I need to rebuild the path name...
				// FQP := fmt.Sprintf("%s/%s", path, f.Name())
				// fmt.Printf("%s\n", FQP)

				// uuid
				u4 := uuid.NewV4()
				fileInfo.UUID = u4.String()

				// set type from predicate
				fileInfo.Filetype = predicate
				fileInfo.Filename = f.Name()

				// md5
				data, err := ioutil.ReadFile(fp)
				if err != nil {
					fmt.Println(err)
				}
				fileInfo.MD5 = md5.Sum(data)

				// Corewall Archives (CARs) are too big to index..  skip them.
				iscar, err := filepath.Match(strings.ToLower("*.car"), strings.ToLower(f.Name()))
				dir := ""
				file := ""

				if !iscar {
					// md5
					data, err := ioutil.ReadFile(fp)
					if err != nil {
						fmt.Println(err)
					}
					fileInfo.MD5 = md5.Sum(data)

					// content via Tika
					fmt.Println("Calling Tika...  can take some time")

					url := "http://localhost:9998/tika"
					req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
					req.Header.Set("Accept", "text/plain")

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						fmt.Println(err)
					}
					defer resp.Body.Close()

					fmt.Println("Response Status:", resp.Status)
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						fmt.Println(err)
					}
					fileInfo.ContentNoStopWords = stopwords.CleanString(string(body), "en", true) // .LenContentNoStopWords = utf8.RuneCountInString(cleanBody)
					dir, file = filepath.Split(fp)

					// Bleve indexer
					loopindex, berr := bleve.Open("csdcoFX.bleve")
					berr = loopindex.Index(fileInfo.UUID, fileInfo)
					berr = loopindex.Close()
					if berr != nil {
						fmt.Printf("Bleve error indexing %v \n", berr)
					}
				}
				if iscar {
					fmt.Println("in the iscar file")
					dir, file = filepath.Split(fp)
					var emtpyByte [16]byte
					fileInfo.MD5 = emtpyByte
				}

				//  Build the triples here and then append to the master set
				newsub, _ := rdf.NewIRI(fmt.Sprintf("http://opencoredata/id/resource/csdco/datafile/%s", fileInfo.UUID)) // Sprintf a correct URI here
				newpred0, _ := rdf.NewIRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#type")
				newobj0, _ := rdf.NewIRI("http://opencoredata.org/id/voc/csdco/v1/Datafile")
				newtriple0 := rdf.Triple{Subj: newsub, Pred: newpred0, Obj: newobj0}

				newpred1, _ := rdf.NewIRI("http://opencoredata.org/id/voc/csdco/v1/fileuuid")
				newobj1, _ := rdf.NewLiteral(fileInfo.UUID)
				newtriple1 := rdf.Triple{Subj: newsub, Pred: newpred1, Obj: newobj1}

				// newpred2, _ := rdf.NewIRI("http://opencoredata.org/id/voc/csdco/v1/size")
				// newobj2, _ := rdf.NewLiteral(fileInfo.LenContentNoStopWords)
				// newtriple2 := rdf.Triple{Subj: newsub, Pred: newpred2, Obj: newobj2}

				newpred3, _ := rdf.NewIRI("http://opencoredata.org/id/voc/csdco/v1/name")
				newobj3, _ := rdf.NewLiteral(file)
				newtriple3 := rdf.Triple{Subj: newsub, Pred: newpred3, Obj: newobj3}

				//  Project IRI if I make a match
				projectURI := blazeCall(projectID)
				if projectURI != "" {
					newpred4, err := rdf.NewIRI("http://opencoredata.org/id/voc/csdco/v1/Project")
					if err != nil {
						fmt.Println(err)
					}
					newobj4, err := rdf.NewIRI(projectURI)
					if err != nil {
						fmt.Println(err)
					}
					newtriple4 := rdf.Triple{Subj: newsub, Pred: newpred4, Obj: newobj4}
					tr = append(tr, newtriple4)
				}

				// Project name in as a literal
				newpred4v2, _ := rdf.NewIRI("http://opencoredata.org/id/voc/csdco/v1/ProjectName")
				newobj4v2, _ := rdf.NewLiteral(projectID)
				newtriple4v2 := rdf.Triple{Subj: newsub, Pred: newpred4v2, Obj: newobj4v2}

				newpred5, _ := rdf.NewIRI("http://opencoredata.org/id/voc/csdco/v1/FileType")
				newobj5, _ := rdf.NewIRI(predicate)
				newtriple5 := rdf.Triple{Subj: newsub, Pred: newpred5, Obj: newobj5}

				newpred6, _ := rdf.NewIRI("http://opencoredata.org/id/voc/csdco/v1/FileLocation")
				newobj6, _ := rdf.NewLiteral(fp)
				newtriple6 := rdf.Triple{Subj: newsub, Pred: newpred6, Obj: newobj6}

				// TODO  add in MD5 value
				newpred7, _ := rdf.NewIRI("http://opencoredata.org/id/voc/csdco/v1/FileMD5")
				newobj7, _ := rdf.NewLiteral(fmt.Sprintf("%x", fileInfo.MD5))
				newtriple7 := rdf.Triple{Subj: newsub, Pred: newpred7, Obj: newobj7}

				// will any every need to be skipped to add triples?
				tr = append(tr, newtriple0)
				tr = append(tr, newtriple1)
				// tr = append(tr, newtriple2)
				tr = append(tr, newtriple3)
				tr = append(tr, newtriple4v2)
				tr = append(tr, newtriple5)
				tr = append(tr, newtriple6)
				tr = append(tr, newtriple7)

				// // split so I can use a slice element in a lookup for metadata
				// fmt.Printf("%q\n", strings.Split(path, "/"))

				// I could write each collection of triples to a file at the fileID level here.  As they are .nt
				// I could then just cat them together.  Still, while a good checkpoint concept..  I don't really
				// do any real checkpoint approach here..  perhaps if I add that (like a boltdb with KV values in it
				// of completed files )  I can revist how this tr data is managed.

				fmt.Printf("For path:\t%s \nFor dir:\t%s\nFor file:\t%s\nMD5:\t%x \nUUID:\t%s  \n\n", fp, dir, file, fileInfo.MD5, fileInfo.UUID)

				// write out the file with the proj name  or if blanktime stamped
				// Serialize the triples to a file...
				filename := fmt.Sprintf("./rdf/%striples.nt", u4.String())
				writeFile(filename, tr)

			}

		}
		if err != nil {
			fmt.Println(err)
		}

		return err
	})

	if err != nil {
		fmt.Println(err)
	}

	// Serialize the triples to a file...  original location
	// writeFile("./rdf/indexerTriples.nt", tr)

	return size, err
}

func inApprovedList(projectName string) bool {
	if projectName == "CAHO" {
		return true
	}
	return false
}

func ageInYears(fp string) float64 {
	fi, err := os.Stat(fp)
	if err != nil {
		fmt.Println(err)
	}
	stat := fi.Sys().(*syscall.Stat_t)
	ctime := time.Unix(int64(stat.Ctimespec.Sec), int64(stat.Ctimespec.Nsec))
	delta := time.Now().Sub(ctime)
	years := delta.Hours() / 24 / 365
	// fmt.Printf("Create: %v   making it %.2f  years old\n", ctime, years)
	return years
}

func caseInsenstiveContains(a, b string) bool {
	return strings.Contains(strings.ToUpper(a), strings.ToUpper(b))
}

func writeFile(name string, tr []rdf.Triple) {
	// Create the output file
	outFile, err := os.Create(name)
	if err != nil {
		log.Fatal(err) // why Fatal on creating the file?  If I can't make it I can't make it...
	}
	defer outFile.Close()

	// Write triples to a file
	var inoutFormat rdf.Format
	inoutFormat = rdf.NTriples // Turtle NQuads
	enc := rdf.NewTripleEncoder(outFile, inoutFormat)
	err = enc.EncodeAll(tr)
	// err = enc.Encode(newtriple)
	enc.Close()
	if err != nil {
		log.Fatal(err) // why Fatal on encoding?
	}
}

func blazeCall(project string) string {
	repo, err := sparql.NewRepo("http://localhost:19999/blazegraph/namespace/csdco/sparql")
	// repo, err := sparql.NewRepo("http://opencoredata.org/sparql")

	if err != nil {
		log.Printf("query make repo: %v\n", err)
	}

	f := bytes.NewBufferString(queries)
	bank := sparql.LoadBank(f)

	q, err := bank.Prepare("urionly", project)
	if err != nil {
		log.Printf("query bank prepair: %v\n", err)
	}

	res, err := repo.Query(q)

	if err != nil {
		log.Printf("query call: %v\n", err)
	}

	bindingsTest := res.Results.Bindings // map[string][]rdf.Term
	var URI string
	for _, i := range bindingsTest {
		URI = fmt.Sprintf("%v", i["uri"].Value)
	}

	return URI
}

// visit  Deprectated function
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
		panic(err) // why panic here?
	}
	defer resp.Body.Close()

	// filter out some files that we don't want to index? dot files, what else?
	// filter out stop words and numbers?  No need to index these.
	fmt.Println("Response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	// body, _ := ioutil.ReadAll(resp.Body)
	// cleanBody := stopwords.CleanString(string(body), "en", true)
	// fileInfo.LenContentNoStopWords = utf8.RuneCountInString(cleanBody)
	dir, file := filepath.Split(path)

	// // split so I can use a slice element in a lookup for metadata
	// fmt.Printf("%q\n", strings.Split(path, "/"))

	fmt.Printf("For path:\t%s \nFor dir:\t%s\nFor file:\t%s\nMD5:\t%v \nUUID:\t%s  \n\n", path, dir, file, fileInfo.MD5, fileInfo.UUID)

	//     fmt.Printf("is not dir: %s\n", path)
	// fmt.Printf("Found: %s  %d\n", f.Name(), f.Size())

	// indexJSON()  // pass to Bleve for indexing
	// generateTriples() // build triples and append to a flat file for later use in triple store

	return nil
}
