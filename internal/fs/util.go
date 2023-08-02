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
