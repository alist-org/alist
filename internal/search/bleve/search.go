package bleve

import (
	"context"
	"os"

	query2 "github.com/blevesearch/bleve/v2/search/query"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/errs"
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

func (b *Bleve) Config() searcher.Config {
	return config
}

func (b *Bleve) Search(ctx context.Context, req model.SearchReq) ([]model.SearchNode, int64, error) {
	var queries []query2.Query
	query := bleve.NewMatchQuery(req.Keywords)
	query.SetField("name")
	queries = append(queries, query)
	if req.Scope != 0 {
		isDir := req.Scope == 1
		isDirQuery := bleve.NewBoolFieldQuery(isDir)
		queries = append(queries, isDirQuery)
	}
	reqQuery := bleve.NewConjunctionQuery(queries...)
	search := bleve.NewSearchRequest(reqQuery)
	search.SortBy([]string{"name"})
	search.From = (req.Page - 1) * req.PerPage
	search.Size = req.PerPage
	search.Fields = []string{"*"}
	searchResults, err := b.BIndex.Search(search)
	if err != nil {
		log.Errorf("search error: %+v", err)
		return nil, 0, err
	}
	res, err := utils.SliceConvert(searchResults.Hits, func(src *search2.DocumentMatch) (model.SearchNode, error) {
		return model.SearchNode{
			Parent: src.Fields["parent"].(string),
			Name:   src.Fields["name"].(string),
			IsDir:  src.Fields["is_dir"].(bool),
			Size:   int64(src.Fields["size"].(float64)),
		}, nil
	})
	return res, int64(searchResults.Total), nil
}

func (b *Bleve) Index(ctx context.Context, node model.SearchNode) error {
	return b.BIndex.Index(uuid.NewString(), node)
}

func (b *Bleve) BatchIndex(ctx context.Context, nodes []model.SearchNode) error {
	batch := b.BIndex.NewBatch()
	for _, node := range nodes {
		batch.Index(uuid.NewString(), node)
	}
	return b.BIndex.Batch(batch)
}

func (b *Bleve) Get(ctx context.Context, parent string) ([]model.SearchNode, error) {
	return nil, errs.NotSupport
}

func (b *Bleve) Del(ctx context.Context, prefix string) error {
	return errs.NotSupport
}

func (b *Bleve) Release(ctx context.Context) error {
	if b.BIndex != nil {
		return b.BIndex.Close()
	}
	return nil
}

func (b *Bleve) Clear(ctx context.Context) error {
	err := b.Release(ctx)
	if err != nil {
		return err
	}
	log.Infof("Removing old index...")
	err = os.RemoveAll(conf.Conf.BleveDir)
	if err != nil {
		log.Errorf("clear bleve error: %+v", err)
	}
	bIndex, err := Init(&conf.Conf.BleveDir)
	if err != nil {
		return err
	}
	b.BIndex = bIndex
	return nil
}

var _ searcher.Searcher = (*Bleve)(nil)
