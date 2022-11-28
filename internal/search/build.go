package search

import (
	"context"
	"path"
	"path/filepath"
	"time"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	log "github.com/sirupsen/logrus"
)

var (
	Running = false
)

func BuildIndex(ctx context.Context, indexPaths, ignorePaths []string, maxDepth int, count bool) error {
	var objCount uint64 = 0
	Running = true
	var (
		err error
		fi  model.Obj
	)
	defer func() {
		Running = false
		now := time.Now()
		eMsg := ""
		if err != nil {
			log.Errorf("build index error: %+v", err)
			eMsg = err.Error()
		} else {
			log.Infof("success build index, count: %d", objCount)
		}
		if count {
			WriteProgress(&model.IndexProgress{
				ObjCount:     objCount,
				IsDone:       err == nil,
				LastDoneTime: &now,
				Error:        eMsg,
			})
		}
	}()
	admin, err := db.GetAdmin()
	if err != nil {
		return err
	}
	if count {
		WriteProgress(&model.IndexProgress{
			ObjCount: 0,
			IsDone:   false,
		})
	}
	for _, indexPath := range indexPaths {
		walkFn := func(indexPath string, info model.Obj, err error) error {
			for _, avoidPath := range ignorePaths {
				if indexPath == avoidPath {
					return filepath.SkipDir
				}
			}
			// ignore root
			if indexPath == "/" {
				return nil
			}
			err = Index(ctx, path.Dir(indexPath), info)
			if err != nil {
				return err
			} else {
				objCount++
			}
			if objCount%100 == 0 {
				log.Infof("index obj count: %d", objCount)
				log.Debugf("current success index: %s", indexPath)
				if count {
					WriteProgress(&model.IndexProgress{
						ObjCount:     objCount,
						IsDone:       false,
						LastDoneTime: nil,
					})
				}
			}
			return nil
		}
		fi, err = fs.Get(ctx, indexPath)
		if err != nil {
			return err
		}
		// TODO: run walkFS concurrently
		err = fs.WalkFS(context.WithValue(ctx, "user", admin), maxDepth, indexPath, fi, walkFn)
		if err != nil {
			return err
		}
	}
	return nil
}

func Clear(ctx context.Context) error {
	return instance.Clear(ctx)
}
