package bleve

import (
	"context"
	"path"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/search/searcher"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/blevesearch/bleve/v2"
	search2 "github.com/blevesearch/bleve/v2/search"
	log "github.com/sirupsen/logrus"
)

type Bleve struct {
	BIndex bleve.Index
}

func (b *Bleve) Search(ctx context.Context, req model.SearchReq) ([]model.SearchNode, int64, error) {
	query := bleve.NewMatchQuery(req.Keywords)
	search := bleve.NewSearchRequest(query)
	search.Size = req.PerPage
	search.Fields = []string{"Path"}
	searchResults, err := index.Search(search)
	if err != nil {
		log.Errorf("search error: %+v", err)
		return nil, 0, err
	}
	res, err := utils.SliceConvert(searchResults.Hits, func(src *search2.DocumentMatch) (model.SearchNode, error) {
		p := src.Fields["Path"].(string)
		return model.SearchNode{
			Parent: path.Dir(p),
			Name:   path.Base(p),
		}, nil
	})
	return res, int64(len(res)), nil
}

func (b *Bleve) Index(ctx context.Context, path string, obj model.Obj) error {
	//TODO implement me
	panic("implement me")
}

func (b *Bleve) Del(ctx context.Context, path string, maxDepth int) error {
	//TODO implement me
	panic("implement me")
}

var _ searcher.Searcher = (*Bleve)(nil)

func init() {
	searcher.RegisterDriver(searcher.Config{
		Name: "bleve",
	}, func() searcher.Searcher {
		return nil
	})
}
