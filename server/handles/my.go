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

// 10s自动清理
var listCache = cache.NewMemCache(cache.WithClearInterval[model.DirCache](10 * time.Second))

func MyHandleFsList(path string, objs []model.Obj) {
	// 还没开始添加存储
	if path == "/" && len(objs) == 0 {
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
	// 子文件为 空 的情况
	if len(objs) == 0 {
		if ok && dirCache.Size != 0 {
			dirCache.Size = 0
			db.UpdateDirCache(&dirCache)
			listCache.Set(path, dirCache)
		}
		return
	}

	// 子文件夹包含的 dirCache
	for _, obj := range objs {
		if obj.IsDir() {
			fullPath := stdpath.Join(path, obj.GetName())

			dirCache, ok := lo.Find(dirCaches, func(item model.DirCache) bool {
				return item.Path == fullPath
			})
			if ok {
				listCache.Set(fullPath, dirCache)
			}
		}
	}

	// 获取子文件 最大的修改时间
	maxModifieObj := lo.MaxBy(objs, func(item model.Obj, max model.Obj) bool {
		fullPath := stdpath.Join(path, item.GetName())
		modified := item.ModTime()
		dirCache, ok := MyGetDirCach(fullPath)
		if ok {
			if modified.Before(dirCache.Modified) {
				modified = dirCache.Modified
			}
		}
		return modified.After(max.ModTime())
	})
	// 子文件 大小相加
	sum := lo.Reduce(objs, func(agg int, item model.Obj, _ int) int {
		size := int(item.GetSize())
		if item.IsDir() {
			fullPath := stdpath.Join(path, item.GetName())
			dirCache, ok := MyGetDirCach(fullPath)
			if ok {
				size = int(dirCache.Size)
			}
		}
		return agg + size
	}, 0)

	changedSize := 0
	if ok {
		isChanged := false
		if maxModifieObj.ModTime().After(dirCache.Modified) {
			// 更新父文件夹 修改时间
			dirCache.Modified = maxModifieObj.ModTime()
			isChanged = true
		}
		if dirCache.Size != int64(sum) {
			changedSize = sum - int(dirCache.Size)
			dirCache.Size = int64(sum)
			isChanged = true
		}
		if isChanged {
			if err := db.UpdateDirCache(&dirCache); err != nil {
				utils.Log.Errorf("failed update dirCache: %s", err)
			}
			listCache.Set(path, dirCache)
		}
	} else {
		dirCache = model.DirCache{
			Path:     path,
			Modified: maxModifieObj.ModTime(),
			Size:     int64(sum),
		}
		if err := db.CreateDirCache(&dirCache); err != nil {
			utils.Log.Errorf("failed create dirCache: %s", err)
		}
		changedSize = sum
	}

	// 父文件夹，父父文件夹。。。
	if changedSize > 0 {
		parentFolders := []string{}
		parentPath := path
		for {
			parentPath = stdpath.Dir(parentPath)
			if parentPath == "/" {
				break
			}
			parentFolders = append(parentFolders, parentPath)
		}
		parentCaches, err := db.GetDirCachesByManyPath(parentFolders)
		if err == nil {
			for _, parentCache := range parentCaches {
				parentCache.Size += int64(changedSize)
				if err := db.UpdateDirCache(&parentCache); err != nil {
					utils.Log.Errorf("failed update dirCache: %s", err)
				}
			}
		}
	}
}

func MyGetDirCach(path string) (model.DirCache, bool) {
	return listCache.Get(path)
}
