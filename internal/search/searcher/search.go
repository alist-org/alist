package searcher

import (
	"context"

	"github.com/alist-org/alist/v3/internal/model"
)

type Searcher interface {
	Search(ctx context.Context, req model.SearchReq) ([]model.SearchNode, int64, error)
	Update(ctx context.Context, path string, objs []model.Obj) error
	BuildIndex(ctx context.Context, indexPaths, ignorePaths []string, maxDepth int) error
	Progress(ctx context.Context) (model.IndexProgress, error)
}
