package lanzou

import (
	"context"
	"net/http"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

var upClient = base.NewRestyClient().SetTimeout(120 * time.Second)

type LanZou struct {
	Addition
	model.Storage
}

func (d *LanZou) Config() driver.Config {
	return config
}

func (d *LanZou) GetAddition() driver.Additional {
	return d.Addition
}

func (d *LanZou) Init(ctx context.Context, storage model.Storage) error {
	d.Storage = storage
	err := utils.Json.UnmarshalFromString(d.Storage.Addition, &d.Addition)
	if err != nil {
		return err
	}
	if d.IsCookie() {
		if d.RootFolderID == "" {
			d.RootFolderID = "-1"
		}
	}
	return nil
}

func (d *LanZou) Drop(ctx context.Context) error {
	return nil
}

// 获取的大小和时间不准确
func (d *LanZou) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	if d.IsCookie() {
		return d.GetFiles(ctx, dir.GetID())
	} else {
		return d.GetFileOrFolderByShareUrl(ctx, dir.GetID(), d.SharePassword)
	}
}

func (d *LanZou) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	downID := file.GetID()
	pwd := d.SharePassword
	if d.IsCookie() {
		share, err := d.getFileShareUrlByID(ctx, file.GetID())
		if err != nil {
			return nil, err
		}
		downID = share.FID
		pwd = share.Pwd
	}
	fileInfo, err := d.getFilesByShareUrl(ctx, downID, pwd, nil)
	if err != nil {
		return nil, err
	}

	return &model.Link{
		URL: fileInfo.Url,
		Header: http.Header{
			"User-Agent": []string{base.UserAgent},
		},
	}, nil
}

func (d *LanZou) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	if d.IsCookie() {
		_, err := d.post(d.BaseUrl+"/doupload.php", func(req *resty.Request) {
			req.SetContext(ctx)
			req.SetFormData(map[string]string{
				"task":               "2",
				"parent_id":          parentDir.GetID(),
				"folder_name":        dirName,
				"folder_description": "",
			})
		}, nil)
		return err
	}
	return errs.NotImplement
}

func (d *LanZou) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	if d.IsCookie() {
		if !srcObj.IsDir() {
			_, err := d.post(d.BaseUrl+"/doupload.php", func(req *resty.Request) {
				req.SetContext(ctx)
				req.SetFormData(map[string]string{
					"task":      "20",
					"folder_id": dstDir.GetID(),
					"file_id":   srcObj.GetID(),
				})
			}, nil)
			return err
		}
	}
	return errs.NotImplement
}

func (d *LanZou) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	if d.IsCookie() {
		if !srcObj.IsDir() {
			_, err := d.post(d.BaseUrl+"/doupload.php", func(req *resty.Request) {
				req.SetContext(ctx)
				req.SetFormData(map[string]string{
					"task":      "46",
					"file_id":   srcObj.GetID(),
					"file_name": newName,
					"type":      "2",
				})
			}, nil)
			return err
		}
	}
	return errs.NotImplement
}

func (d *LanZou) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotImplement
}

func (d *LanZou) Remove(ctx context.Context, obj model.Obj) error {
	if d.IsCookie() {
		_, err := d.post(d.BaseUrl+"/doupload.php", func(req *resty.Request) {
			req.SetContext(ctx)
			if obj.IsDir() {
				req.SetFormData(map[string]string{
					"task":      "3",
					"folder_id": obj.GetID(),
				})
			} else {
				req.SetFormData(map[string]string{
					"task":    "6",
					"file_id": obj.GetID(),
				})
			}
		}, nil)
		return err
	}
	return errs.NotImplement
}

func (d *LanZou) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	if d.IsCookie() {
		_, err := d._post(d.BaseUrl+"/fileup.php", func(req *resty.Request) {
			req.SetFormData(map[string]string{
				"task":      "1",
				"id":        "WU_FILE_0",
				"name":      stream.GetName(),
				"folder_id": dstDir.GetID(),
			}).SetFileReader("upload_file", stream.GetName(), stream)
		}, nil, true)
		return err
	}
	return errs.NotImplement
}
