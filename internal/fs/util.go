package fs

import (
	"io"
	"mime"
	"net/http"
	"os"
	stdpath "path"
	"strings"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

func containsByName(files []model.FileInfo, file model.FileInfo) bool {
	for _, f := range files {
		if f.GetName() == file.GetName() {
			return true
		}
	}
	return false
}

var httpClient = &http.Client{}

func getFileStreamFromLink(file model.FileInfo, link *model.Link) (model.FileStreamer, error) {
	var rc io.ReadCloser
	mimetype := mime.TypeByExtension(stdpath.Ext(file.GetName()))
	if link.Data != nil {
		rc = link.Data
	} else if link.FilePath != nil {
		f, err := os.Open(*link.FilePath)
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
	stream := model.FileStream{
		FileInfo:   file,
		ReadCloser: rc,
		Mimetype:   mimetype,
	}
	return stream, nil
}
