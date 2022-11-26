package searcher

import (
	"context"

	"github.com/alist-org/alist/v3/internal/model"
)

type Config struct {
	Name       string
	AutoUpdate bool
}

type Searcher interface {
	// Search specific keywords in specific path
	Search(ctx context.Context, req model.SearchReq) ([]model.SearchNode, int64, error)
	// Index obj with path
	Index(ctx context.Context, parent string, obj model.Obj) error
	// Del path
	Del(ctx context.Context, path string, maxDepth int) error
	// Drop searcher, clear all index now
	Drop(ctx context.Context) error
}
