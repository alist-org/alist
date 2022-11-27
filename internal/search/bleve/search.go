package bleve

import (
	"context"
	"os"
	"path"

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
	query := bleve.NewMatchQuery(req.Keywords)
	query.SetField("Path")
	search := bleve.NewSearchRequest(query)
	search.Size = req.PerPage
	search.Fields = []string{"*"}
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
			IsDir:  src.Fields["IsDir"].(bool),
			Size:   src.Fields["Size"].(int64),
		}, nil
	})
	return res, int64(len(res)), nil
}

type Data struct {
	Path  string
	IsDir bool
	Size  int64
}

func (b *Bleve) Index(ctx context.Context, parent string, obj model.Obj) error {
	return b.BIndex.Index(uuid.NewString(), Data{
		Path:  path.Join(parent, obj.GetName()),
		IsDir: obj.IsDir(),
		Size:  obj.GetSize(),
	})
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
	log.Infof("Removing old index...")
	err := os.RemoveAll(conf.Conf.BleveDir)
	if err != nil {
		log.Errorf("clear bleve error: %+v", err)
	}
	return nil
}

var _ searcher.Searcher = (*Bleve)(nil)
