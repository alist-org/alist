package onedrive_app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sync"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

type OnedriveAPP struct {
	model.Storage
	Addition
	AccessToken string
	root        *Object
	mutex       sync.Mutex
}

func (d *OnedriveAPP) Config() driver.Config {
	return config
}

func (d *OnedriveAPP) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *OnedriveAPP) Init(ctx context.Context) error {
	if d.ChunkSize < 1 {
		d.ChunkSize = 5
	}
	return d.accessToken()
}

func (d *OnedriveAPP) Drop(ctx context.Context) error {
	return nil
}

func (d *OnedriveAPP) GetRoot(ctx context.Context) (model.Obj, error) {
	if d.root != nil {
		return d.root, nil
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	root := &Object{
		ObjThumb: model.ObjThumb{
			Object: model.Object{
				ID:       "root",
				Path:     d.RootFolderPath,
				Name:     "root",
				Size:     0,
				Modified: d.Modified,
				Ctime:    d.Modified,
				IsFolder: true,
			},
		},
		ParentID: "",
	}
	if !utils.PathEqual(d.RootFolderPath, "/") {
		// get root folder id
		url := d.GetMetaUrl(false, d.RootFolderPath)
		var resp struct {
			Id string `json:"id"`
		}
		_, err := d.Request(url, http.MethodGet, nil, &resp)
		if err != nil {
			return nil, err
		}
		root.ID = resp.Id
	}
	d.root = root
	return d.root, nil
}

func (d *OnedriveAPP) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.getFiles(dir.GetPath())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return fileToObj(src, dir.GetID()), nil
	})
}

func (d *OnedriveAPP) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	f, err := d.GetFile(file.GetPath())
	if err != nil {
		return nil, err
	}
	if f.File == nil {
		return nil, errs.NotFile
	}
	u := f.Url
	if d.CustomHost != "" {
		_u, err := url.Parse(f.Url)
		if err != nil {
			return nil, err
		}
		_u.Host = d.CustomHost
		u = _u.String()
	}
	return &model.Link{
		URL: u,
	}, nil
}

func (d *OnedriveAPP) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	url := d.GetMetaUrl(false, parentDir.GetPath()) + "/children"
	data := base.Json{
		"name":                              dirName,
		"folder":                            base.Json{},
		"@microsoft.graph.conflictBehavior": "rename",
	}
	_, err := d.Request(url, http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *OnedriveAPP) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	parentPath := ""
	if dstDir.GetID() == "" {
		parentPath = dstDir.GetPath()
		if utils.PathEqual(parentPath, "/") {
			parentPath = path.Join("/drive/root", parentPath)
		} else {
			parentPath = path.Join("/drive/root:/", parentPath)
		}
	}
	data := base.Json{
		"parentReference": base.Json{
			"id":   dstDir.GetID(),
			"path": parentPath,
		},
		"name": srcObj.GetName(),
	}
	url := d.GetMetaUrl(false, srcObj.GetPath())
	_, err := d.Request(url, http.MethodPatch, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *OnedriveAPP) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	var parentID string
	if o, ok := srcObj.(*Object); ok {
		parentID = o.ParentID
	} else {
		return fmt.Errorf("srcObj is not Object")
	}
	if parentID == "" {
		parentID = "root"
	}
	data := base.Json{
		"parentReference": base.Json{
			"id": parentID,
		},
		"name": newName,
	}
	url := d.GetMetaUrl(false, srcObj.GetPath())
	_, err := d.Request(url, http.MethodPatch, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *OnedriveAPP) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	dst, err := d.GetFile(dstDir.GetPath())
	if err != nil {
		return err
	}
	data := base.Json{
		"parentReference": base.Json{
			"driveId": dst.ParentReference.DriveId,
			"id":      dst.Id,
		},
		"name": srcObj.GetName(),
	}
	url := d.GetMetaUrl(false, srcObj.GetPath()) + "/copy"
	_, err = d.Request(url, http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *OnedriveAPP) Remove(ctx context.Context, obj model.Obj) error {
	url := d.GetMetaUrl(false, obj.GetPath())
	_, err := d.Request(url, http.MethodDelete, nil, nil)
	return err
}

func (d *OnedriveAPP) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	var err error
	if stream.GetSize() <= 4*1024*1024 {
		err = d.upSmall(ctx, dstDir, stream)
	} else {
		err = d.upBig(ctx, dstDir, stream, up)
	}
	return err
}

var _ driver.Driver = (*OnedriveAPP)(nil)
