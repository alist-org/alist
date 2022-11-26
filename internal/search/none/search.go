package none

import (
	"context"

	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/search/searcher"
)

type None struct{}

func (n None) Search(ctx context.Context, req model.SearchReq) ([]model.SearchNode, int64, error) {
	return nil, 0, errs.SearchNotAvailable
}

func (n None) Index(ctx context.Context, path string, obj model.Obj) error {
	return errs.SearchNotAvailable
}

func (n None) Del(ctx context.Context, path string, maxDepth int) error {
	return errs.SearchNotAvailable
}

func init() {
	searcher.RegisterDriver(searcher.Config{Name: "none"}, func() searcher.Searcher {
		return None{}
	})
}
