package fs

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	stdpath "path"
	"strings"
)

func getFileStreamFromLink(file model.Obj, link *model.Link) (*model.FileStream, error) {
	var rc io.ReadCloser
	mimetype := utils.GetMimeType(file.GetName())
	if link.Data != nil {
		rc = link.Data
	} else if link.FilePath != nil {
		// create a new temp symbolic link, because it will be deleted after upload
		newFilePath := stdpath.Join(conf.Conf.TempDir, fmt.Sprintf("%s-%s", uuid.NewString(), file.GetName()))
		err := utils.SymlinkOrCopyFile(*link.FilePath, newFilePath)
		if err != nil {
			return nil, err
		}
		f, err := os.Open(newFilePath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open file %s", *link.FilePath)
		}
		rc = f
	} else if link.Writer != nil {
		r, w := io.Pipe()
		go func() {
			err := link.Writer(w)
			err = w.CloseWithError(err)
			if err != nil {
				log.Errorf("[getFileStreamFromLink] failed to write: %v", err)
			}
		}()
		rc = r
	} else {
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

func isFileCached(path string) bool {
	log.Debugf("isFileCached %s", path)
	return utils.SliceContains(cachedFiles, path)
}
func syncCache(main driver.Driver, path string, cache driver.Driver) uint64 {
	//using CopyTaskManager but should have it own taskManager, I dont know web
	return CopyTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64]{
		Name: fmt.Sprintf("copy [%s](%s) to [%s](%s)", main.GetStorage().MountPath, path, cache.GetStorage().MountPath, path),
		Func: func(t *task.Task[uint64]) error {
			actualPath := op.GetAndActualWithStorage(path, main)
			err := op.Remove(t.Ctx, cache, actualPath)
			if err != nil {
				return err
			}
			err = copyFileBetween2Storages(t, main, cache, actualPath, stdpath.Dir(actualPath))
			if err == nil {
				fifoCache(t.Ctx, cache, path)
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
