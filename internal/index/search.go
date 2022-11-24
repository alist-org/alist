package index

import (
	"github.com/blevesearch/bleve/v2"
	log "github.com/sirupsen/logrus"
)

func Search(queryString string, size int) (*bleve.SearchResult, error) {
	query := bleve.NewMatchQuery(queryString)
	search := bleve.NewSearchRequest(query)
	search.Size = size
	search.Fields = []string{"Path"}
	searchResults, err := index.Search(search)
	if err != nil {
		log.Errorf("search error: %+v", err)
		return nil, err
	}
	return searchResults, nil
}
