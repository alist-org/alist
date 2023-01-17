package db_non_full_text

import (
	"context"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/search/searcher"
)

type DB struct{}

func (D DB) Config() searcher.Config {
	return config
}

func (D DB) Search(ctx context.Context, req model.SearchReq) ([]model.SearchNode, int64, error) {
	return db.SearchNode(req, false)
}

func (D DB) Index(ctx context.Context, node model.SearchNode) error {
	return db.CreateSearchNode(&node)
}

func (D DB) BatchIndex(ctx context.Context, nodes []model.SearchNode) error {
	return db.BatchCreateSearchNodes(&nodes)
}

func (D DB) Get(ctx context.Context, parent string) ([]model.SearchNode, error) {
	return db.GetSearchNodesByParent(parent)
}

func (D DB) Del(ctx context.Context, path string) error {
	return db.DeleteSearchNodesByParent(path)
}

func (D DB) Release(ctx context.Context) error {
	return nil
}

func (D DB) Clear(ctx context.Context) error {
	return db.ClearSearchNodes()
}

var _ searcher.Searcher = (*DB)(nil)
