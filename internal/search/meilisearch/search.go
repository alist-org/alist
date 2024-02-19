package meilisearch

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/search/searcher"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"strings"
)

type searchDocument struct {
	ID string `json:"id"`
	model.SearchNode
}

type Meilisearch struct {
	Client               *meilisearch.Client
	IndexUid             string
	FilterableAttributes []string
	SearchableAttributes []string
}

func (m *Meilisearch) Config() searcher.Config {
	return config
}

func (m *Meilisearch) Search(ctx context.Context, req model.SearchReq) ([]model.SearchNode, int64, error) {
	mReq := &meilisearch.SearchRequest{
		AttributesToSearchOn: m.SearchableAttributes,
		Page:                 int64(req.Page),
		HitsPerPage:          int64(req.PerPage),
	}
	if req.Scope != 0 {
		mReq.Filter = fmt.Sprintf("is_dir = %v", req.Scope == 1)
	}
	search, err := m.Client.Index(m.IndexUid).Search(req.Keywords, mReq)
	if err != nil {
		return nil, 0, err
	}
	nodes, err := utils.SliceConvert(search.Hits, func(src any) (model.SearchNode, error) {
		srcMap := src.(map[string]any)
		return model.SearchNode{
			Parent: srcMap["parent"].(string),
			Name:   srcMap["name"].(string),
			IsDir:  srcMap["is_dir"].(bool),
			Size:   int64(srcMap["size"].(float64)),
		}, nil
	})
	if err != nil {
		return nil, 0, err
	}
	return nodes, search.TotalHits, nil
}

func (m *Meilisearch) Index(ctx context.Context, node model.SearchNode) error {
	return m.BatchIndex(ctx, []model.SearchNode{node})
}

func (m *Meilisearch) BatchIndex(ctx context.Context, nodes []model.SearchNode) error {
	documents, _ := utils.SliceConvert(nodes, func(src model.SearchNode) (*searchDocument, error) {

		return &searchDocument{
			ID:         uuid.NewString(),
			SearchNode: src,
		}, nil
	})

	_, err := m.Client.Index(m.IndexUid).AddDocuments(documents)
	if err != nil {
		return err
	}

	//// Wait for the task to complete and check
	//forTask, err := m.Client.WaitForTask(task.TaskUID, meilisearch.WaitParams{
	//	Context:  ctx,
	//	Interval: time.Millisecond * 50,
	//})
	//if err != nil {
	//	return err
	//}
	//if forTask.Status != meilisearch.TaskStatusSucceeded {
	//	return fmt.Errorf("BatchIndex failed, task status is %s", forTask.Status)
	//}
	return nil
}

func (m *Meilisearch) Get(ctx context.Context, parent string) ([]model.SearchNode, error) {
	var result meilisearch.DocumentsResult
	err := m.Client.Index(m.IndexUid).GetDocuments(&meilisearch.DocumentsQuery{
		Filter: fmt.Sprintf("parent = '%s'", strings.ReplaceAll(parent, "'", "\\'")),
		Limit:  int64(model.MaxInt),
	}, &result)
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(result.Results, func(src map[string]any) (model.SearchNode, error) {
		return model.SearchNode{
			Parent: src["parent"].(string),
			Name:   src["name"].(string),
			IsDir:  src["is_dir"].(bool),
			Size:   int64(src["size"].(float64)),
		}, nil
	})

}

func (m *Meilisearch) Del(ctx context.Context, prefix string) error {
	return errs.NotSupport
}

func (m *Meilisearch) Release(ctx context.Context) error {
	return nil
}

func (m *Meilisearch) Clear(ctx context.Context) error {
	_, err := m.Client.Index(m.IndexUid).DeleteAllDocuments()
	return err
}
