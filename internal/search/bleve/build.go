package bleve

import (
	"context"
	"path/filepath"
	"time"

	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/blevesearch/bleve/v2"
	"github.com/google/uuid"
)

type Data struct {
	Path string
}

func BuildIndex(ctx context.Context, indexPaths, ignorePaths []string, maxDepth int) {
	// TODO: partial remove indices
	Reset()
	var batchs []*bleve.Batch
	var fileCount uint64 = 0
	for _, indexPath := range indexPaths {
		batch := func() *bleve.Batch {
			batch := index.NewBatch()
			// TODO: cache unchanged part
			walkFn := func(indexPath string, info model.Obj, err error) error {
				for _, avoidPath := range ignorePaths {
					if indexPath == avoidPath {
						return filepath.SkipDir
					}
				}
				if !info.IsDir() {
					batch.Index(uuid.NewString(), Data{Path: indexPath})
					fileCount += 1
					if fileCount%100 == 0 {
						WriteProgress(&model.IndexProgress{
							FileCount:    fileCount,
							IsDone:       false,
							LastDoneTime: nil,
						})
					}
				}
				return nil
			}
			fi, err := fs.Get(ctx, indexPath)
			if err != nil {
				return batch
			}
			// TODO: run walkFS concurrently
			fs.WalkFS(ctx, maxDepth, indexPath, fi, walkFn)
			return batch
		}()
		batchs = append(batchs, batch)
	}
	for _, batch := range batchs {
		index.Batch(batch)
	}
	now := time.Now()
	WriteProgress(&model.IndexProgress{
		FileCount:    fileCount,
		IsDone:       true,
		LastDoneTime: &now,
	})
}
