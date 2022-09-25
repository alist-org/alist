package aliyundrive_share

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/cron"
	"github.com/alist-org/alist/v3/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type AliyundriveShare struct {
	model.Storage
	Addition
	AccessToken string
	ShareToken  string
	DriveId     string
	cron        *cron.Cron
}

func (d *AliyundriveShare) Config() driver.Config {
	return config
}

func (d *AliyundriveShare) GetAddition() driver.Additional {
	return d.Addition
}

func (d *AliyundriveShare) Init(ctx context.Context, storage model.Storage) error {
	d.Storage = storage
	err := utils.Json.UnmarshalFromString(d.Storage.Addition, &d.Addition)
	if err != nil {
		return err
	}
	err = d.refreshToken()
	if err != nil {
		return err
	}
	err = d.getShareToken()
	if err != nil {
		return err
	}
	d.cron = cron.NewCron(time.Hour * 2)
	d.cron.Do(func() {
		err := d.refreshToken()
		if err != nil {
			log.Errorf("%+v", err)
		}
	})
	return nil
}

func (d *AliyundriveShare) Drop(ctx context.Context) error {
	if d.cron != nil {
		d.cron.Stop()
	}
	return nil
}

func (d *AliyundriveShare) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.getFiles(dir.GetID())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return fileToObj(src), nil
	})
}

//func (d *AliyundriveShare) Get(ctx context.Context, path string) (model.Obj, error) {
//	// this is optional
//	return nil, errs.NotImplement
//}

func (d *AliyundriveShare) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	data := base.Json{
		"drive_id":   d.DriveId,
		"file_id":    file.GetID(),
		"expire_sec": 14400,
	}
	var e ErrorResp
	res, err := base.RestyClient.R().
		SetError(&e).SetBody(data).
		SetHeader("content-type", "application/json").
		SetHeader("Authorization", "Bearer\t"+d.AccessToken).
		Post("https://api.aliyundrive.com/v2/file/get_download_url")
	if err != nil {
		return nil, err
	}
	var u string
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" {
			err = d.refreshToken()
			if err != nil {
				return nil, err
			}
			return d.Link(ctx, file, args)
		} else if e.Code == "ForbiddenNoPermission.File" {
			data = utils.MergeMap(data, base.Json{
				// Only ten minutes valid
				"expire_sec": 600,
				"share_id":   d.ShareId,
			})
			var resp ShareLinkResp
			var e2 ErrorResp
			_, err = base.RestyClient.R().
				SetError(&e2).SetBody(data).SetResult(&resp).
				SetHeader("content-type", "application/json").
				SetHeader("Authorization", "Bearer\t"+d.AccessToken).
				SetHeader("x-share-token", d.ShareToken).
				Post("https://api.aliyundrive.com/v2/file/get_share_link_download_url")
			if err != nil {
				return nil, err
			}
			if e2.Code != "" {
				if e2.Code == "AccessTokenInvalid" || e2.Code == "ShareLinkTokenInvalid" {
					err = d.getShareToken()
					if err != nil {
						return nil, err
					}
					return d.Link(ctx, file, args)
				} else {
					return nil, errors.New(e2.Code + ":" + e2.Message)
				}
			} else {
				u = resp.DownloadUrl
			}
		} else {
			return nil, errors.New(e.Code + ":" + e.Message)
		}
	} else {
		u = utils.Json.Get(res.Body(), "url").ToString()
	}
	return &model.Link{
		Header: http.Header{
			"Referer": []string{"https://www.aliyundrive.com/"},
		},
		URL: u,
	}, nil
}

func (d *AliyundriveShare) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	// TODO create folder
	return errs.NotSupport
}

func (d *AliyundriveShare) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO move obj
	return errs.NotSupport
}

func (d *AliyundriveShare) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	// TODO rename obj
	return errs.NotSupport
}

func (d *AliyundriveShare) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj
	return errs.NotSupport
}

func (d *AliyundriveShare) Remove(ctx context.Context, obj model.Obj) error {
	// TODO remove obj
	return errs.NotSupport
}

func (d *AliyundriveShare) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// TODO upload file
	return errs.NotSupport
}

var _ driver.Driver = (*AliyundriveShare)(nil)
