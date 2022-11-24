package index

import (
	"context"
	"path"
	"path/filepath"
	"time"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/blevesearch/bleve/v2"
	"github.com/google/uuid"
)

// walkFS traverses filesystem fs starting at name up to depth levels.
//
// walkFS will stop when current depth > `depth`. For each visited node,
// walkFS calls walkFn. If a visited file system node is a directory and
// walkFn returns path.SkipDir, walkFS will skip traversal of this node.
func walkFS(ctx context.Context, depth int, name string, info model.Obj, walkFn func(reqPath string, info model.Obj, err error) error) error {
	// This implementation is based on Walk's code in the standard path/path package.
	walkFnErr := walkFn(name, info, nil)
	if walkFnErr != nil {
		if info.IsDir() && walkFnErr == filepath.SkipDir {
			return nil
		}
		return walkFnErr
	}
	if !info.IsDir() || depth == 0 {
		return nil
	}
	meta, _ := db.GetNearestMeta(name)
	// Read directory names.
	objs, err := fs.List(context.WithValue(ctx, "meta", meta), name)
	if err != nil {
		return walkFnErr
	}
	for _, fileInfo := range objs {
		filename := path.Join(name, fileInfo.GetName())
		if err := walkFS(ctx, depth-1, filename, fileInfo, walkFn); err != nil {
			if err == filepath.SkipDir {
				break
			}
			return err
		}
	}
	return nil
}

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
						WriteProgress(&Progress{
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
			walkFS(ctx, maxDepth, indexPath, fi, walkFn)
			return batch
		}()
		batchs = append(batchs, batch)
	}
	for _, batch := range batchs {
		index.Batch(batch)
	}
	now := time.Now()
	WriteProgress(&Progress{
		FileCount:    fileCount,
		IsDone:       true,
		LastDoneTime: &now,
	})
}
