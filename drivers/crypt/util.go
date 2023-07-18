package crypt

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/net"
	"github.com/alist-org/alist/v3/pkg/utils"
	"net/http"
	stdpath "path"
	"strconv"
)

func (d *Crypt) list(ctx context.Context, dst, sub string) ([]model.Obj, error) {
	objs, err := fs.List(ctx, stdpath.Join(dst, sub), &fs.ListArgs{NoLog: true})
	// the obj must implement the model.SetPath interface
	// return objs, err
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(objs, func(obj model.Obj) (model.Obj, error) {
		thumb, ok := model.GetThumb(obj)
		objRes := model.Object{
			Name:     obj.GetName(),
			Size:     obj.GetSize(),
			Modified: obj.ModTime(),
			IsFolder: obj.IsDir(),
		}
		if !ok {
			return &objRes, nil
		}
		return &model.ObjThumb{
			Object: objRes,
			Thumbnail: model.Thumbnail{
				Thumbnail: thumb,
			},
		}, nil
	})
}

func RequestRangedHttp(r *http.Request, link *model.Link, offset, length int64) (*http.Response, error) {
	header := net.ProcessHeader(&http.Header{}, &link.Header)
	if offset == 0 && length < 0 {
		header.Del("Range")
	} else {
		end := ""
		if length >= 0 {
			end = strconv.FormatInt(offset+length-1, 10)
		}
		header.Set("Range", fmt.Sprintf("bytes=%v-%v", offset, end))
	}

	return net.RequestHttp("GET", header, link.URL)
}
