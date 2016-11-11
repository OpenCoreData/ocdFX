package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/knakk/sparql"
)

const queries = `
# Comments are ignored, except those tagging a query.

# tag: ocdFX1
SELECT ?uuid ?name ?location
WHERE 
{ 
  ?uri  <http://opencoredata.org/id/voc/csdco/v1/fileuuid> ?uuid .
  ?uri  <http://opencoredata.org/id/voc/csdco/v1/name> ?name .
  ?uri  <http://opencoredata.org/id/voc/csdco/v1/FileLocation> ?location .
 }
`

func main() {
	// setup bolt if it is not already
	setupBolt()

	// open Bolt DB
	db, err := bolt.Open("catalog.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// make sparql call
	repo, err := sparql.NewRepo("http://localhost:9999/blazegraph/namespace/ocdfx/sparql") // repo, err := sparql.NewRepo("http://opencoredata.org/sparql")
	if err != nil {
		log.Printf("query make repo: %v\n", err)
	}

	f := bytes.NewBufferString(queries)
	bank := sparql.LoadBank(f)
	q, err := bank.Prepare("ocdFX1")
	if err != nil {
		log.Printf("query bank prepair: %v\n", err)
	}

	res, err := repo.Query(q)
	if err != nil {
		log.Printf("query call: %v\n", err)
	}

	// Print loop testing
	bindingsTest := res.Results.Bindings // []map[string][]rdf.Term
	fmt.Println("res.Results.Bindings:")
	for k, i := range bindingsTest {
		fmt.Printf("At postion %v with %v and %v  and %v\n\n", k, i["uuid"].Value, i["name"].Value, i["location"].Value)

		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("Bucket1"))
			err := b.Put([]byte(i["uuid"].Value), []byte(i["location"].Value))
			return err
		})

	}
}

func setupBolt() {

	db, err := bolt.Open("catalog.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// You can also create a bucket only if it doesn't exist by using the Tx.CreateBucketIfNotExists()
	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("Bucket1"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		log.Printf("Bucket created %v", b.FillPercent)
		return nil
	})
}
