package search

import (
	"context"
	"path"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/mq"
	"github.com/alist-org/alist/v3/pkg/utils"
	log "github.com/sirupsen/logrus"
)

var (
	Running = atomic.Bool{}
	Quit    chan struct{}
)

func BuildIndex(ctx context.Context, indexPaths, ignorePaths []string, maxDepth int, count bool) error {
	storages, err := db.GetEnabledStorages()
	if err != nil {
		return err
	}
	var skipDrivers = []string{"AList V2", "AList V3"}
	for _, storage := range storages {
		if utils.SliceContains(skipDrivers, storage.Driver) {
			// TODO: request for indexing permission
			ignorePaths = append(ignorePaths, storage.MountPath)
		}
	}
	var (
		objCount uint64 = 0
		fi       model.Obj
	)
	Running.Store(true)
	Quit = make(chan struct{}, 1)
	indexMQ := mq.NewInMemoryMQ[ObjWithParent]()
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				log.Infof("index obj count: %d", objCount)
				indexMQ.ConsumeAll(func(messages []mq.Message[ObjWithParent]) {
					if len(messages) != 0 {
						log.Debugf("current index: %s", messages[len(messages)-1].Content.Parent)
					}
					if err = BatchIndex(ctx, utils.MustSliceConvert(messages,
						func(src mq.Message[ObjWithParent]) ObjWithParent {
							return src.Content
						})); err != nil {
						log.Errorf("build index in batch error: %+v", err)
					} else {
						objCount = objCount + uint64(len(messages))
					}
					if count {
						WriteProgress(&model.IndexProgress{
							ObjCount:     objCount,
							IsDone:       false,
							LastDoneTime: nil,
						})
					}
				})

			case <-Quit:
				Running.Store(false)
				ticker.Stop()
				eMsg := ""
				now := time.Now()
				originErr := err
				indexMQ.ConsumeAll(func(messages []mq.Message[ObjWithParent]) {
					if err = BatchIndex(ctx, utils.MustSliceConvert(messages,
						func(src mq.Message[ObjWithParent]) ObjWithParent {
							return src.Content
						})); err != nil {
						log.Errorf("build index in batch error: %+v", err)
					} else {
						objCount = objCount + uint64(len(messages))
					}
					if originErr != nil {
						log.Errorf("build index error: %+v", err)
						eMsg = err.Error()
					} else {
						log.Infof("success build index, count: %d", objCount)
					}
					if count {
						WriteProgress(&model.IndexProgress{
							ObjCount:     objCount,
							IsDone:       true,
							LastDoneTime: &now,
							Error:        eMsg,
						})
					}
				})
				return
			}
		}
	}()
	defer func() {
		if Running.Load() {
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
			if !Running.Load() {
				return filepath.SkipDir
			}
			for _, avoidPath := range ignorePaths {
				if strings.HasPrefix(indexPath, avoidPath) {
					return filepath.SkipDir
				}
			}
			// ignore root
			if indexPath == "/" {
				return nil
			}
			indexMQ.Publish(mq.Message[ObjWithParent]{
				Content: ObjWithParent{
					Obj:    info,
					Parent: path.Dir(indexPath),
				},
			})
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
