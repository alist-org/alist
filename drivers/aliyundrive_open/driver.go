package aliyundrive_open

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Xhofe/rateg"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

type AliyundriveOpen struct {
	model.Storage
	Addition
	base string

	DriveId string

	limitList func(ctx context.Context, data base.Json) (*Files, error)
	limitLink func(ctx context.Context, file model.Obj) (*model.Link, error)
}

func (d *AliyundriveOpen) Config() driver.Config {
	return config
}

func (d *AliyundriveOpen) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *AliyundriveOpen) Init(ctx context.Context) error {
	if d.LIVPDownloadFormat == "" {
		d.LIVPDownloadFormat = "jpeg"
	}
	if d.DriveType == "" {
		d.DriveType = "default"
	}
	res, err := d.request("/adrive/v1.0/user/getDriveInfo", http.MethodPost, nil)
	if err != nil {
		return err
	}
	d.DriveId = utils.Json.Get(res, d.DriveType+"_drive_id").ToString()
	d.limitList = rateg.LimitFnCtx(d.list, rateg.LimitFnOption{
		Limit:  4,
		Bucket: 1,
	})
	d.limitLink = rateg.LimitFnCtx(d.link, rateg.LimitFnOption{
		Limit:  1,
		Bucket: 1,
	})
	return nil
}

func (d *AliyundriveOpen) Drop(ctx context.Context) error {
	return nil
}

func (d *AliyundriveOpen) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	if d.limitList == nil {
		return nil, fmt.Errorf("driver not init")
	}
	files, err := d.getFiles(ctx, dir.GetID())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return fileToObj(src), nil
	})
}

func (d *AliyundriveOpen) link(ctx context.Context, file model.Obj) (*model.Link, error) {
	res, err := d.request("/adrive/v1.0/openFile/getDownloadUrl", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"drive_id":   d.DriveId,
			"file_id":    file.GetID(),
			"expire_sec": 14400,
		})
	})
	if err != nil {
		return nil, err
	}
	url := utils.Json.Get(res, "url").ToString()
	if url == "" {
		if utils.Ext(file.GetName()) != "livp" {
			return nil, errors.New("get download url failed: " + string(res))
		}
		url = utils.Json.Get(res, "streamsUrl", d.LIVPDownloadFormat).ToString()
	}
	exp := time.Minute
	return &model.Link{
		URL:        url,
		Expiration: &exp,
	}, nil
}

func (d *AliyundriveOpen) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	if d.limitLink == nil {
		return nil, fmt.Errorf("driver not init")
	}
	return d.limitLink(ctx, file)
}

func (d *AliyundriveOpen) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error) {
	nowTime, _ := getNowTime()
	newDir := File{CreatedAt: nowTime, UpdatedAt: nowTime}
	_, err := d.request("/adrive/v1.0/openFile/create", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"drive_id":        d.DriveId,
			"parent_file_id":  parentDir.GetID(),
			"name":            dirName,
			"type":            "folder",
			"check_name_mode": "refuse",
		}).SetResult(&newDir)
	})
	if err != nil {
		return nil, err
	}
	return fileToObj(newDir), nil
}

func (d *AliyundriveOpen) Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	var resp MoveOrCopyResp
	_, err := d.request("/adrive/v1.0/openFile/move", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"drive_id":          d.DriveId,
			"file_id":           srcObj.GetID(),
			"to_parent_file_id": dstDir.GetID(),
			"check_name_mode":   "refuse", // optional:ignore,auto_rename,refuse
			//"new_name":          "newName", // The new name to use when a file of the same name exists
		}).SetResult(&resp)
	})
	if err != nil {
		return nil, err
	}
	if resp.Exist {
		return nil, errors.New("existence of files with the same name")
	}

	if srcObj, ok := srcObj.(*model.ObjThumb); ok {
		srcObj.ID = resp.FileID
		srcObj.Modified = time.Now()
		return srcObj, nil
	}
	return nil, nil
}

func (d *AliyundriveOpen) Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error) {
	var newFile File
	_, err := d.request("/adrive/v1.0/openFile/update", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"drive_id": d.DriveId,
			"file_id":  srcObj.GetID(),
			"name":     newName,
		}).SetResult(&newFile)
	})
	if err != nil {
		return nil, err
	}
	return fileToObj(newFile), nil
}

func (d *AliyundriveOpen) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	_, err := d.request("/adrive/v1.0/openFile/copy", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"drive_id":          d.DriveId,
			"file_id":           srcObj.GetID(),
			"to_parent_file_id": dstDir.GetID(),
			"auto_rename":       true,
		})
	})
	return err
}

func (d *AliyundriveOpen) Remove(ctx context.Context, obj model.Obj) error {
	uri := "/adrive/v1.0/openFile/recyclebin/trash"
	if d.RemoveWay == "delete" {
		uri = "/adrive/v1.0/openFile/delete"
	}
	_, err := d.request(uri, http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"drive_id": d.DriveId,
			"file_id":  obj.GetID(),
		})
	})
	return err
}

func (d *AliyundriveOpen) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	return d.upload(ctx, dstDir, stream, up)
}

func (d *AliyundriveOpen) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
	var resp base.Json
	var uri string
	data := base.Json{
		"drive_id": d.DriveId,
		"file_id":  args.Obj.GetID(),
	}
	switch args.Method {
	case "video_preview":
		uri = "/adrive/v1.0/openFile/getVideoPreviewPlayInfo"
		data["category"] = "live_transcoding"
		data["url_expire_sec"] = 14400
	default:
		return nil, errs.NotSupport
	}
	_, err := d.request(uri, http.MethodPost, func(req *resty.Request) {
		req.SetBody(data).SetResult(&resp)
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

var _ driver.Driver = (*AliyundriveOpen)(nil)
var _ driver.MkdirResult = (*AliyundriveOpen)(nil)
var _ driver.MoveResult = (*AliyundriveOpen)(nil)
var _ driver.RenameResult = (*AliyundriveOpen)(nil)
var _ driver.PutResult = (*AliyundriveOpen)(nil)
