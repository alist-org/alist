package search

import (
	"context"
	"path"
	"path/filepath"
	"time"

	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
)

func BuildIndex(ctx context.Context, indexPaths, ignorePaths []string, maxDepth int) error {
	var objCount uint64 = 0
	var (
		err error
		fi  model.Obj
	)
	defer func() {
		now := time.Now()
		eMsg := ""
		if err != nil {
			eMsg = err.Error()
		}
		WriteProgress(&model.IndexProgress{
			FileCount:    objCount,
			IsDone:       err == nil,
			LastDoneTime: &now,
			Error:        eMsg,
		})
	}()
	for _, indexPath := range indexPaths {
		walkFn := func(indexPath string, info model.Obj, err error) error {
			for _, avoidPath := range ignorePaths {
				if indexPath == avoidPath {
					return filepath.SkipDir
				}
			}
			err = instance.Index(ctx, path.Dir(indexPath), info)
			if err != nil {
				return err
			}
			if objCount%100 == 0 {
				WriteProgress(&model.IndexProgress{
					FileCount:    objCount,
					IsDone:       false,
					LastDoneTime: nil,
				})
			}
			return nil
		}
		fi, err = fs.Get(ctx, indexPath)
		if err != nil {
			return err
		}
		// TODO: run walkFS concurrently
		err = fs.WalkFS(ctx, maxDepth, indexPath, fi, walkFn)
		if err != nil {
			return err
		}
	}
	return nil
}
