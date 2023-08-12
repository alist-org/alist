package fs

import (
	"github.com/alist-org/alist/v3/pkg/http_range"
	"io"
	"net/http"
	"strings"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/pkg/errors"
)

func getFileStreamFromLink(file model.Obj, link *model.Link) (*model.FileStream, error) {
	var rc io.ReadCloser
	var err error
	mimetype := utils.GetMimeType(file.GetName())
	if link.RangeReadCloser.RangeReader != nil {
		rc, err = link.RangeReadCloser.RangeReader(http_range.Range{Length: -1})
		if err != nil {
			return nil, err
		}
	} else if link.ReadSeekCloser != nil {
		rc = link.ReadSeekCloser
	} else {
		//TODO: add accelerator
		req, err := http.NewRequest(http.MethodGet, link.URL, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create request for %s", link.URL)
		}
		for h, val := range link.Header {
			req.Header[h] = val
		}
		res, err := common.HttpClient().Do(req)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get response for %s", link.URL)
		}
		mt := res.Header.Get("Content-Type")
		if mt != "" && strings.ToLower(mt) != "application/octet-stream" {
			mimetype = mt
		}
		rc = res.Body
	}
	// if can't get mimetype, use default application/octet-stream
	if mimetype == "" {
		mimetype = "application/octet-stream"
	}
	stream := &model.FileStream{
		Obj:        file,
		ReadCloser: rc,
		Mimetype:   mimetype,
	}
	return stream, nil
}

var cachedFiles []string

func chooseCache(storage driver.Driver, path string) driver.Driver {
	cacheStorage, has := op.HasCacheStorage(storage.GetStorage().MountPath)
	if has {
		if isFileCached(path) {
			storage = cacheStorage
		} else {
			syncCache(storage, path, cacheStorage)
		}
	}
	return storage
}
func isFileCached(path string) bool {
	ret := utils.SliceContains(cachedFiles, path)
	log.Debugf("isFileCached %s: %t", path, ret)
	return ret

}
func syncCache(main driver.Driver, path string, cache driver.Driver) uint64 {
	//using CopyTaskManager but should have it own taskManager, I dont know web
	return CopyTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64]{
		Name: fmt.Sprintf("cache [%s](%s) to [%s](%s)", main.GetStorage().MountPath, path, cache.GetStorage().MountPath, path),
		Func: func(t *task.Task[uint64]) error {
			actualPath := op.GetAndActualWithStorage(path, main)

			srcFile, err := op.Get(t.Ctx, main, actualPath)
			// only cache file exclude dir
			if (err == nil) && !(srcFile.IsDir()) {

				err = op.Remove(t.Ctx, cache, actualPath)
				if err != nil {
					return err
				}
				err = copyFileBetween2Storages(t, main, cache, actualPath, stdpath.Dir(actualPath))
				if err == nil {
					fifoCache(t.Ctx, cache, path)
				}
			}
			log.Debugf("copy file between storages: %+v", err)
			return err
		},
	}))

}

// LRU better than fifo, use database to do
// restart means drop and overflow
func fifoCache(ctx context.Context, storage driver.Driver, path string) {
	// what should i do if the task fail?
	if len(cachedFiles) <= 100 {
		cachedFiles = append(cachedFiles, path)
	}

	if len(cachedFiles) > 100 {
		actualPath := op.GetAndActualWithStorage(path, storage)
		// Cache will lag behind main deletion when fifo is full
		err := op.Remove(ctx, storage, actualPath)
		if err == nil {
			cachedFiles = cachedFiles[1:100]
		}
	}

}
