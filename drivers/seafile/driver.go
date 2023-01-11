package seafile

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

type Seafile struct {
	model.Storage
	Addition

	authorization string
}

func (d *Seafile) Config() driver.Config {
	return config
}

func (d *Seafile) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Seafile) Init(ctx context.Context) error {
	d.Address = strings.TrimSuffix(d.Address, "/")
	return d.getToken()
}

func (d *Seafile) Drop(ctx context.Context) error {
	return nil
}

func (d *Seafile) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	path := dir.GetPath()
	var resp []RepoDirItemResp
	_, err := d.request(http.MethodGet, fmt.Sprintf("/api2/repos/%s/dir/", d.Addition.RepoId), func(req *resty.Request) {
		req.SetResult(&resp).SetQueryParams(map[string]string{
			"p": path,
		})
	})
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(resp, func(f RepoDirItemResp) (model.Obj, error) {
		return &model.ObjThumb{
			Object: model.Object{
				Name:     f.Name,
				Modified: time.Unix(f.Modified, 0),
				Size:     f.Size,
				IsFolder: f.Type == "dir",
			},
			// Thumbnail: model.Thumbnail{Thumbnail: f.Thumb},
		}, nil
	})
}

func (d *Seafile) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	res, err := d.request(http.MethodGet, fmt.Sprintf("/api2/repos/%s/file/", d.Addition.RepoId), func(req *resty.Request) {
		req.SetQueryParams(map[string]string{
			"p":     file.GetPath(),
			"reuse": "1",
		})
	})
	if err != nil {
		return nil, err
	}
	u := string(res)
	u = u[1 : len(u)-1] // remove quotes
	return &model.Link{URL: u}, nil
}

func (d *Seafile) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	_, err := d.request(http.MethodPost, fmt.Sprintf("/api2/repos/%s/dir/", d.Addition.RepoId), func(req *resty.Request) {
		req.SetQueryParams(map[string]string{
			"p": filepath.Join(parentDir.GetPath(), dirName),
		}).SetFormData(map[string]string{
			"operation": "mkdir",
		})
	})
	return err
}

func (d *Seafile) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	_, err := d.request(http.MethodPost, fmt.Sprintf("/api2/repos/%s/file/", d.Addition.RepoId), func(req *resty.Request) {
		req.SetQueryParams(map[string]string{
			"p": srcObj.GetPath(),
		}).SetFormData(map[string]string{
			"operation": "move",
			"dst_repo":  d.Addition.RepoId,
			"dst_dir":   dstDir.GetPath(),
		})
	}, true)
	return err
}

func (d *Seafile) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	_, err := d.request(http.MethodPost, fmt.Sprintf("/api2/repos/%s/file/", d.Addition.RepoId), func(req *resty.Request) {
		req.SetQueryParams(map[string]string{
			"p": srcObj.GetPath(),
		}).SetFormData(map[string]string{
			"operation": "rename",
			"newname":   newName,
		})
	}, true)
	return err
}

func (d *Seafile) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	_, err := d.request(http.MethodPost, fmt.Sprintf("/api2/repos/%s/file/", d.Addition.RepoId), func(req *resty.Request) {
		req.SetQueryParams(map[string]string{
			"p": srcObj.GetPath(),
		}).SetFormData(map[string]string{
			"operation": "copy",
			"dst_repo":  d.Addition.RepoId,
			"dst_dir":   dstDir.GetPath(),
		})
	})
	return err
}

func (d *Seafile) Remove(ctx context.Context, obj model.Obj) error {
	_, err := d.request(http.MethodDelete, fmt.Sprintf("/api2/repos/%s/file/", d.Addition.RepoId), func(req *resty.Request) {
		req.SetQueryParams(map[string]string{
			"p": obj.GetPath(),
		})
	})
	return err
}

func (d *Seafile) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	res, err := d.request(http.MethodGet, fmt.Sprintf("/api2/repos/%s/upload-link/", d.Addition.RepoId), func(req *resty.Request) {
		req.SetQueryParams(map[string]string{
			"p": dstDir.GetPath(),
		})
	})
	if err != nil {
		return err
	}

	u := string(res)
	u = u[1 : len(u)-1] // remove quotes
	_, err = d.request(http.MethodPost, u, func(req *resty.Request) {
		req.SetFileReader("file", stream.GetName(), stream).
			SetFormData(map[string]string{
				"parent_dir": dstDir.GetPath(),
				"replace":    "1",
			})
	})
	return err
}

var _ driver.Driver = (*Seafile)(nil)
