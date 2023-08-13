package fs

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/generic_sync"
	"github.com/alist-org/alist/v3/pkg/task"
	//try groupcache or database
	"github.com/hashicorp/golang-lru/v2"
	log "github.com/sirupsen/logrus"
	stdpath "path"
	"time"
)

var cachedFiles = generic_sync.MapOf[driver.Driver, *lru.Cache[string, time.Time]]{}

func chooseCache(ctx context.Context, storage driver.Driver, path string) driver.Driver {
	cacheStorage, has := op.HasCacheStorage(storage.GetStorage().MountPath)
	if has {
		if isFileCached(storage, path) && cacheUpdated(ctx, storage, cacheStorage, path) {
			storage = cacheStorage
		} else {
			//cache reocrd cant use, so remove
			removeCacheRecord(storage, path)
			copy2Cache(storage, cacheStorage, path)
		}
	}
	return storage
}

func isFileCached(main driver.Driver, path string) bool {
	lruCache, _ := cachedFiles.Load(main)
	_, ret := lruCache.Get(path)
	log.Debugf("isFileCached %s: %t", path, ret)
	return ret

}
func cacheUpdated(ctx context.Context, main driver.Driver, cache driver.Driver, path string) bool {
	//TODO
	mainActualpath := op.GetActualWithStorage(path, main)
	mainSrcFile, mainErr := op.Get(ctx, main, mainActualpath)
	//get modTime from cacheDatabase instead of storage
	//cacheActualpath := op.GetActualWithStorage(path, cache)
	//cacheSrcFile, cacheErr := op.Get(ctx, cache, cacheActualpath)

	lruCache, _ := cachedFiles.Load(main)
	cacheModTime, _ := lruCache.Get(path)
	if mainErr == nil {
		ret := !cacheModTime.Before(mainSrcFile.ModTime())
		log.Debugf("cacheUpdated %s : %t", path, ret)
		return ret
	}
	//return if cache newer than main
	return false
}

// copy to cache storage and then record
func copy2Cache(main driver.Driver, cache driver.Driver, path string) uint64 {
	//using CopyTaskManager but should have it own taskManager, I dont know web
	return CopyTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64]{
		Name: fmt.Sprintf("cache [%s](%s) to [%s](%s)", main.GetStorage().MountPath, path, cache.GetStorage().MountPath, path),
		Func: func(t *task.Task[uint64]) error {
			mainActualpath := op.GetActualWithStorage(path, main)

			mainSrcFile, err := op.Get(t.Ctx, main, mainActualpath)
			// only cache file exclude dir
			if (err == nil) && !(mainSrcFile.IsDir()) {

				err = op.Remove(t.Ctx, cache, mainActualpath)
				if err != nil {
					log.Debugf("remove stale file fail: %+v", err)
				}
				err = copyFileBetween2Storages(t, main, cache, mainActualpath, stdpath.Dir(mainActualpath))
				if err == nil {
					appendCacheRecord(t.Ctx, main, path, mainSrcFile.ModTime())
				}
			}
			return err
		},
	}))

}
func removeCacheRecord(main driver.Driver, path string) {

}

// LRU better than fifo, use database to do
// restart means drop and overflow
func appendCacheRecord(ctx context.Context, main driver.Driver, path string, modTime time.Time) {
	CacheSize := 100
	// what should i do if the task fail?
	lruCache, ok := cachedFiles.Load(main)
	if !ok {
		lruCache, _ = lru.New[string, time.Time](CacheSize)
		cachedFiles.Store(main, lruCache)
	}
	if lruCache.Len() < CacheSize {

		lruCache.Add(path, modTime)
	}

	if lruCache.Len() >= CacheSize {
		cacheStorage, has := op.HasCacheStorage(main.GetStorage().MountPath)
		if has {
			k, _, _ := lruCache.GetOldest()
			lruCache.Remove(k)
			err := op.Remove(ctx, cacheStorage, k)

			if err != nil {
				log.Debugf("remove obsolote fail :%+v", err)
			}
		}
	}

}
