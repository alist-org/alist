package base

import (
	"io"
	"net/http"
	"strconv"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/http_range"
	"github.com/alist-org/alist/v3/pkg/utils"
)

func HandleRange(link *model.Link, file io.ReadSeekCloser, header http.Header, size int64) {
	if header.Get("Range") != "" {
		r, err := http_range.ParseRange(header.Get("Range"), size)
		if err == nil && len(r) > 0 {
			_, err := file.Seek(r[0].Start, io.SeekStart)
			if err == nil {
				link.Data = utils.NewLimitReadCloser(file, func() error {
					return file.Close()
				}, r[0].Length)
				link.Status = http.StatusPartialContent
				link.Header = http.Header{
					"Content-Range":  []string{r[0].ContentRange(size)},
					"Content-Length": []string{strconv.FormatInt(r[0].Length, 10)},
				}
			}
		}
	}
}
