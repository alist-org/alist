package aliyundrive_open

import (
	"context"
	"io"
	"math"
	"net/http"
	"strings"

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
}

func (d *AliyundriveOpen) Config() driver.Config {
	return config
}

func (d *AliyundriveOpen) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *AliyundriveOpen) Init(ctx context.Context) error {
	res, err := d.request("/adrive/v1.0/user/getDriveInfo", http.MethodPost, nil)
	if err != nil {
		return err
	}
	d.DriveId = utils.Json.Get(res, "default_drive_id").ToString()
	return nil
}

func (d *AliyundriveOpen) Drop(ctx context.Context) error {
	return nil
}

func (d *AliyundriveOpen) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.getFiles(dir.GetID())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return fileToObj(src), nil
	})
}

func (d *AliyundriveOpen) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
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
	return &model.Link{
		URL: url,
	}, nil
}

func (d *AliyundriveOpen) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	_, err := d.request("/adrive/v1.0/openFile/create", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"drive_id":        d.DriveId,
			"parent_file_id":  parentDir.GetID(),
			"name":            dirName,
			"type":            "folder",
			"check_name_mode": "refuse",
		})
	})
	return err
}

func (d *AliyundriveOpen) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	_, err := d.request("/adrive/v1.0/openFile/move", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"drive_id":          d.DriveId,
			"file_id":           srcObj.GetID(),
			"to_parent_file_id": dstDir.GetID(),
			"check_name_mode":   "refuse", // optional:ignore,auto_rename,refuse
			//"new_name":          "newName", // The new name to use when a file of the same name exists
		})
	})
	return err
}

func (d *AliyundriveOpen) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	_, err := d.request("/adrive/v1.0/openFile/update", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"drive_id": d.DriveId,
			"file_id":  srcObj.GetID(),
			"name":     newName,
		})
	})
	return err
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

func (d *AliyundriveOpen) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// rapid_upload is not currently supported
	// 1. create
	const DEFAULT int64 = 20971520
	createData := base.Json{
		"drive_id":        d.DriveId,
		"parent_file_id":  dstDir.GetID(),
		"name":            stream.GetName(),
		"type":            "file",
		"check_name_mode": "ignore",
	}
	count := 1
	if stream.GetSize() > DEFAULT {
		count = int(math.Ceil(float64(stream.GetSize()) / float64(DEFAULT)))
		partInfoList := make([]base.Json, 0, count)
		for i := 1; i <= count; i++ {
			partInfoList = append(partInfoList, base.Json{"part_number": i})
		}
		createData["part_info_list"] = partInfoList
	}
	var createResp CreateResp
	_, err := d.request("/adrive/v1.0/openFile/create", http.MethodPost, func(req *resty.Request) {
		req.SetBody(createData).SetResult(&createResp)
	})
	if err != nil {
		return err
	}
	// 2. upload
	for i, partInfo := range createResp.PartInfoList {
		if utils.IsCanceled(ctx) {
			return ctx.Err()
		}
		uploadUrl := partInfo.UploadUrl
		if d.InternalUpload {
			//Replace a known public Host with an internal Host
			uploadUrl = strings.ReplaceAll(uploadUrl, "https://cn-beijing-data.aliyundrive.net/", "http://ccp-bj29-bj-1592982087.oss-cn-beijing-internal.aliyuncs.com/")
		}
		req, err := http.NewRequest("PUT", uploadUrl, io.LimitReader(stream, DEFAULT))
		if err != nil {
			return err
		}
		req = req.WithContext(ctx)
		res, err := base.HttpClient.Do(req)
		if err != nil {
			return err
		}
		res.Body.Close()
		if count > 0 {
			up(i * 100 / count)
		}
	}
	// 3. complete
	_, err = d.request("/adrive/v1.0/openFile/complete", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"drive_id":  d.DriveId,
			"file_id":   createResp.FileId,
			"upload_id": createResp.UploadId,
		})
	})
	return err
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
