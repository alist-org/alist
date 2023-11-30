package handles

import (
	stdpath "path"
	"time"

	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/samber/lo"
)

// 20s自动清理
var listCache = cache.NewMemCache(cache.WithClearInterval[model.DirCache](20 * time.Second))

func MyHandleFsList(path string, objs []model.Obj) {
	if len(objs) == 0 {
		return
	}

	dirCaches, err := db.GetDirCachesByPath(path)
	if err != nil {
		return
	}
	// 如果存在
	dirCache, ok := lo.Find(dirCaches, func(item model.DirCache) bool {
		return item.Path == path
	})
	if ok {
		listCache.Set(path, dirCache)
	}

	// 子文件夹
	for _, obj := range objs {
		if obj.IsDir() {
			fullPath := stdpath.Join(path, obj.GetName())

			dirCache, ok := lo.Find(dirCaches, func(item model.DirCache) bool {
				return item.Path == fullPath
			})
			if ok {
				if obj.ModTime().Before(dirCache.Modified) {
					listCache.Set(fullPath, dirCache)
				}
			}
		}
	}

	// 获取子文件夹最大的 修改时间
	maxModifieObj := lo.MaxBy(objs, func(item model.Obj, max model.Obj) bool {
		modified := item.ModTime()
		dirCache, ok := MyGetDirCach(stdpath.Join(path, item.GetName()))
		if ok {
			if modified.Before(dirCache.Modified) {
				modified = dirCache.Modified
			}
		}
		return modified.After(max.ModTime())
	})
	dirCache, ok = lo.Find(dirCaches, func(item model.DirCache) bool {
		return item.Path == path
	})

	// 更新父文件夹 大小
	sum := lo.Reduce(objs, func(agg int, item model.Obj, _ int) int {
		size := MyGetDirCacheSize(path)
		if size == 0 {
			size = int(item.GetSize())
		}
		return agg + size
	}, 0)

	if ok {
		if maxModifieObj.ModTime().After(dirCache.Modified) {
			// 更新父文件夹 修改时间
			dirCache.Modified = maxModifieObj.ModTime()

			dirCache.Size = int64(sum)
			db.UpdateDirCache(&dirCache)
		}
	} else {
		dirCache = model.DirCache{
			Path:     path,
			Modified: maxModifieObj.ModTime(),
			Size:     int64(sum),
		}
		if err := db.CreateDirCache(&dirCache); err != nil {
			utils.Log.Error("failed create dirCache", err)
		}
	}
}

func MyGetDirCacheSize(path string) int {
	dirCache, ok := listCache.Get(path)
	if ok {
		return int(dirCache.Size)
	}
	return 0
}

func MyGetDirCach(path string) (model.DirCache, bool) {
	return listCache.Get(path)
}
