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
	Search(ctx context.Context, req model.SearchReq) ([]model.SearchNode, int64, error)
	Index(ctx context.Context, path string, obj model.Obj) error
	Del(ctx context.Context, path string, maxDepth int) error
}
