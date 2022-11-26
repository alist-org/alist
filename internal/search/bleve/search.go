package bleve

import (
	"context"
	"path"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/search/searcher"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/blevesearch/bleve/v2"
	search2 "github.com/blevesearch/bleve/v2/search"
	"github.com/google/uuid"
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
	searchResults, err := b.BIndex.Search(search)
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

type Data struct {
	Path string
}

func (b *Bleve) Index(ctx context.Context, parent string, obj model.Obj) error {
	return b.BIndex.Index(uuid.NewString(), Data{Path: path.Join(parent, obj.GetName())})
}

func (b *Bleve) Del(ctx context.Context, path string, maxDepth int) error {
	//TODO implement me
	panic("implement me")
}

func (b *Bleve) Drop(ctx context.Context) error {
	if b.BIndex != nil {
		return b.BIndex.Close()
	}
	Reset()
	return nil
}

var _ searcher.Searcher = (*Bleve)(nil)

func init() {
	searcher.RegisterSearcher(searcher.Config{
		Name: "bleve",
	}, func() searcher.Searcher {
		return nil
	})
}
