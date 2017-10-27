package main

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"log"
	"github.com/blevesearch/bleve/search"
)

type doc struct {
	Name       string
	Recipients []string
}

func main() {
	mapping := bleve.NewIndexMapping()
	index, err := bleve.NewMemOnly(mapping)

	mustMatch := []string{"a@example.com", "b@example.com"}

	d := doc{Name: "xyz", Recipients: mustMatch}
	if err := index.Index("1234", d); err != nil {
		log.Fatal(err)
	}

	d2 := doc{Name: "abc", Recipients: []string{"a@example.com", "b@example.com", "c@example.com"}}
	if err := index.Index("5678", d2); err != nil {
		log.Fatal(err)
	}

	query := bleve.NewMatchQuery("a@example.com,b@example.com")
	search := bleve.NewSearchRequest(query)
	search.Fields = []string{"Name", "Recipients"}
	searchResults, err := index.Search(search)
	if err != nil {
		log.Fatal(err)
	}
	bleveResultChildMustMatchArrayFilter(searchResults, "Recipients", mustMatch)
	fmt.Printf("num results: %d\n", searchResults.Hits.Len())

	hit := searchResults.Hits[0]
	fmt.Printf("%s\n", hit.Fields)
}

//
// XXX: HACK - See blevesearch/bleve#637, blevesearch/bleve#15
//
// Currently there is an open issue with Bleve whereby there is no way
// to filter search results to only matching arrays. The result will
// return results that match ANY item in the provided search.
// This function will take an existing search result and then
// filter each hit to ensure the `fieldName` child matches the `mustMatch` array.
func bleveResultChildMustMatchArrayFilter(searchResult *bleve.SearchResult, fieldName string, mustMatch interface{}) {
	// Process the array first
	var mustMatchArr = make([]interface{}, 0)
	if mustMatchStrArr, ok := mustMatch.([]string); ok {
		for _, s := range mustMatchStrArr {
			mustMatchArr = append(mustMatchArr, s)
		}
	} else {
		if iArr, ok := mustMatch.([]interface{}); ok {
			for _, i := range iArr {
				mustMatchArr = append(mustMatchArr, i)
			}
		} else {
			mustMatchArr = append(mustMatchArr, mustMatch)
		}
	}
	newCollection := search.DocumentMatchCollection{}

	for _, c := range searchResult.Hits {
		if v, ok := c.Fields[fieldName]; ok {
			// Quick exit - check is array and length matches
			if vArr, ok := v.([]interface{}); ok && len(vArr) == len(mustMatchArr) {
				// Check that values match
				for i := 0; i < len(vArr); i++ {
					if vArr[i] != mustMatchArr[i] {
						continue
					}
				}

				// Add to collection
				newCollection = append(newCollection, c)
			}
		}
	}

	searchResult.Hits = newCollection
}