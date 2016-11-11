package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

type myServer struct {
	r *mux.Router
}

// Test curl: curl http://localhost:3000/static/csdcofile/97e482ad-0945-4441-a317-723a1fbd56f5
func main() {
	servroute := mux.NewRouter()
	servroute.HandleFunc("/static/csdcofile/{ID}", downloadHandler)
	http.Handle("/static/", servroute)

	// Start the server...
	log.Printf("About to listen on 3000. Go to http://127.0.0.1:3000/")

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Println(vars["ID"])

	// data, err := ioutil.ReadFile(r.URL.Path[1:])
	data, err := ioutil.ReadFile("./static/test.txt")
	log.Println(r.URL.Path[1:])
	if err != nil {
		fmt.Fprint(w, err)
	}

	path := pathFromUUID(vars["ID"])
	dir, file := filepath.Split(path)
	ext := filepath.Ext(path)
	log.Printf("input: %q\n\tdir: %q\n\tfile: %q\text: %q\n", path, dir, file, ext)

	// Pull the extension from the filename and see what we can do
	if ext != "" {
		w.Header().Set("Content-Type", mime.TypeByExtension(ext))
	}

	http.ServeContent(w, r, file, time.Now(), bytes.NewReader(data))
}

func pathFromUUID(UUID string) string {
	db, err := bolt.Open("../LookUpBuilder/catalog.db", 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var v []byte

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Bucket1"))
		v = b.Get([]byte(UUID))
		return nil
	})

	return string(v)
}

func (s *myServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	rw.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	s.r.ServeHTTP(rw, req)
}
