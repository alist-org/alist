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
	Quit    chan struct{}
)

func BuildIndex(ctx context.Context, indexPaths, ignorePaths []string, maxDepth int, count bool) error {
	storages, err := db.GetEnabledStorages()
	if err != nil {
		return err
	}
	for _, storage := range storages {
		if storage.Driver == "AList V2" || storage.Driver == "AList V3" {
			// TODO: request for indexing permission
			ignorePaths = append(ignorePaths, storage.MountPath)
		}
	}
	var (
		objCount uint64 = 0
		fi       model.Obj
	)
	Running = true
	Quit = make(chan struct{}, 1)
	parents := []string{}
	infos := []model.Obj{}
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				log.Infof("index obj count: %d", objCount)
				if len(parents) != 0 {
					log.Debugf("current index: %s", parents[len(parents)-1])
				}
				if err = BatchIndex(ctx, parents, infos); err != nil {
					log.Errorf("build index in batch error: %+v", err)
				} else {
					objCount = objCount + uint64(len(parents))
				}
				if count {
					WriteProgress(&model.IndexProgress{
						ObjCount:     objCount,
						IsDone:       false,
						LastDoneTime: nil,
					})
				}
				parents = nil
				infos = nil
			case <-Quit:
				Running = false
				ticker.Stop()
				eMsg := ""
				now := time.Now()
				originErr := err
				if err = BatchIndex(ctx, parents, infos); err != nil {
					log.Errorf("build index in batch error: %+v", err)
				} else {
					objCount = objCount + uint64(len(parents))
				}
				parents = nil
				infos = nil
				if originErr != nil {
					log.Errorf("build index error: %+v", err)
					eMsg = err.Error()
				} else {
					log.Infof("success build index, count: %d", objCount)
				}
				if count {
					WriteProgress(&model.IndexProgress{
						ObjCount:     objCount,
						IsDone:       originErr == nil,
						LastDoneTime: &now,
						Error:        eMsg,
					})
				}
				return
			}
		}
	}()
	defer func() {
		if Running {
			Quit <- struct{}{}
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
		walkFn := func(indexPath string, info model.Obj) error {
			for _, avoidPath := range ignorePaths {
				if indexPath == avoidPath {
					return filepath.SkipDir
				}
			}
			// ignore root
			if indexPath == "/" {
				return nil
			}
			parents = append(parents, path.Dir(indexPath))
			infos = append(infos, info)
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
