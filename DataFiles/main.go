package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	sparql "github.com/knakk/sparql"
)

const queries = `
# Comments are ignored, except those tagging a query.

# tag: projcall
SELECT DISTINCT *
WHERE {
   ?uri rdf:type <http://opencoredata.org/id/voc/csdco/v1/CSDCOProject> .
   ?uri <http://opencoredata.org/id/voc/csdco/v1/project> "{{.Proj}}" .
   ?uri ?p ?o .
}
`

func main() {
	content, err := ioutil.ReadFile("testSet.txt")
	if err != nil {
		fmt.Printf("Error with ioutils %s\n", err)
	}
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		splits := strings.Split(line, " ")

		fmt.Println(strings.TrimSpace(splits[0]))

		// need to make a SPARQL call for each and pull back the results...
		// res := SPARQLCall(strings.TrimSpace(splits[0]), "projcall", "http://localhost:9999/blazegraph/namespace/csdco/sparql")

		res := CallHack("age")

		fmt.Println(res)

		// solutionsTest := res.Solutions() // map[string][]rdf.Term
		// for _, i := range solutionsTest {
		// 	fmt.Println(i)
		// }
	}
}

func CallHack(age string) string {

	// http://localhost:9999/blazegraph/namespace/csdco/sparql?query=SELECT DISTINCT * WHERE { ?uri rdf:type <http://opencoredata.org/id/voc/csdco/v1/CSDCOProject> . ?uri <http://opencoredata.org/id/voc/csdco/v1/project> "AAFBLP" . ?uri ?p ?o . }

	//  < this & >

	// SELECT%20DISTINCT%20*%20WHERE%20%7B%20%3Furi%20rdf%3Atype%20%3Chttp%3A%2F%2Fopencoredata.org%2Fid%2Fvoc%2Fcsdco%2Fv1%2FCSDCOProject%3E%20.%20%3Furi%20%3Chttp%3A%2F%2Fopencoredata.org%2Fid%2Fvoc%2Fcsdco%2Fv1%2Fproject%3E%20%22AAFBLP%22%20.%20%3Furi%20%3Fp%20%3Fo%20.%20%7D

	const url = "http://localhost:9999/bigdata/namespace/csdco/sparql?query=SELECT%20DISTINCT%20*%20WHERE%20%7B%20%3Furi%20rdf%3Atype%20%3Chttp%3A%2F%2Fopencoredata.org%2Fid%2Fvoc%2Fcsdco%2Fv1%2FCSDCOProject%3E%20.%20%3Furi%20%3Chttp%3A%2F%2Fopencoredata.org%2Fid%2Fvoc%2Fcsdco%2Fv1%2Fproject%3E%20%22AAFBLP%22%20.%20%3Furi%20%3Fp%20%3Fo%20.%20%7D"

	req, _ := http.NewRequest("GET", url, nil)
	// req.Header.Add("accept", "application/sparql-results+json")
	req.Header.Add("accept", "text/csv")
	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	return string(body)

	// dt, err := template.New("RDF template").Parse(url)
	// if err != nil {
	// 	log.Printf("RDF template creation failed for hole data: %s", err)
	// }

	// var buff = bytes.NewBufferString("")
	// err = dt.Execute(buff, age)
	// if err != nil {
	// 	log.Printf("RDF template execution failed: %s", err)
	// }

	// req, _ := http.NewRequest("GET", string(buff.Bytes()), nil)

	// res, _ := http.DefaultClient.Do(req)

	// defer res.Body.Close()
	// body, _ := ioutil.ReadAll(res.Body)

	// csiroStruct := CSIROStruct{}
	// json.Unmarshal(body, &csiroStruct)

	// // loop on csiroStruct.Results.Bindings
	// for _, item := range csiroStruct.Results.Bindings {
	// 	fmt.Println(item.Era.Type)
	// 	fmt.Println(item.Era.Value)
	// 	fmt.Println(item.Name.Type)
	// 	fmt.Println(item.Name.Value)
	// }

	// fmt.Println(string(body))

}

// SPARQLCall ref janusCSVtoGraph (there is a SPAQL call there)
func SPARQLCall(project string, query string, endpoint string) *sparql.Results {
	repo, err := sparql.NewRepo(endpoint,
		sparql.Timeout(time.Millisecond*15000),
	)
	if err != nil {
		log.Printf("Error 1 %s\n", err)
	}

	f := bytes.NewBufferString(queries)
	bank := sparql.LoadBank(f)

	q, err := bank.Prepare(query, struct{ Proj string }{project})
	if err != nil {
		log.Printf("Error 2 %s\n", err)
	}

	fmt.Println(q)

	//qv2 := "SELECT%20DISTINCT%20*%20WHERE%20%7B%20%3Furi%20rdf%3Atype%20%3Chttp%3A%2F%2Fopencoredata.org%2Fid%2Fvoc%2Fcsdco%2Fv1%2FCSDCOProject%3E%20.%20%3Furi%20%3Chttp%3A%2F%2Fopencoredata.org%2Fid%2Fvoc%2Fcsdco%2Fv1%2Fproject%3E%20%22AAFBLP%22%20.%20%3Furi%20%3Fp%20%3Fo%20.%20%7D"

	res, err := repo.Query(q)
	if err != nil {
		log.Printf("Error 3 %s\n", err)
	}

	return res
}
