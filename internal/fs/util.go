package fs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	stdpath "path"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func ClearCache(path string) {
	storage, actualPath, err := op.GetStorageAndActualPath(path)
	if err != nil {
		return
	}
	op.ClearCache(storage, actualPath)
}

func containsByName(files []model.Obj, file model.Obj) bool {
	for _, f := range files {
		if f.GetName() == file.GetName() {
			return true
		}
	}
	return false
}

var httpClient = &http.Client{}

func getFileStreamFromLink(file model.Obj, link *model.Link) (model.FileStreamer, error) {
	var rc io.ReadCloser
	mimetype := utils.GetMimeType(file.GetName())
	if link.Data != nil {
		rc = link.Data
	} else if link.FilePath != nil {
		// copy a new temp, because will be deleted after upload
		newFilePath := stdpath.Join(conf.Conf.TempDir, fmt.Sprintf("%s-%s", uuid.NewString(), file.GetName()))
		err := utils.CopyFile(*link.FilePath, newFilePath)
		if err != nil {
			return nil, err
		}
		f, err := os.Open(newFilePath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open file %s", *link.FilePath)
		}
		rc = f
	} else {
		req, err := http.NewRequest(http.MethodGet, link.URL, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create request for %s", link.URL)
		}
		for h, val := range link.Header {
			req.Header[h] = val
		}
		res, err := httpClient.Do(req)
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
