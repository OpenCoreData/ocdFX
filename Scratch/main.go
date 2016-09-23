package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {

	url := "http://localhost:9999/bigdata/namespace/csdco/sparql?query=SELECT%2520DISTINCT%2520*%2520WHERE%2520%257B%2520%253Furi%2520rdf%253Atype%2520%253Chttp%253A%252F%252Fopencoredata.org%252Fid%252Fvoc%252Fcsdco%252Fv1%252FCSDCOProject%253E%2520.%2520%253Furi%2520%253Chttp%253A%252F%252Fopencoredata.org%252Fid%252Fvoc%252Fcsdco%252Fv1%252Fproject%253E%2520%2522AAFBLP%2522%2520.%2520%253Furi%2520%253Fp%2520%253Fo%2520.%2520%257D"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/sparql-results+json")
	req.Header.Add("cache-control", "no-cache")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))

}
