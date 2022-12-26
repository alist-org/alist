package aliyundrive_share

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
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
	return &d.Addition
}

func (d *AliyundriveShare) Init(ctx context.Context) error {
	err := d.refreshToken()
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

func (d *AliyundriveShare) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	data := base.Json{
		"drive_id": d.DriveId,
		"file_id":  file.GetID(),
		// // Only ten minutes lifetime
		"expire_sec": 600,
		"share_id":   d.ShareId,
	}
	var resp ShareLinkResp
	var e ErrorResp
	_, err := base.RestyClient.R().
		SetError(&e).SetBody(data).SetResult(&resp).
		SetHeader("content-type", "application/json").
		SetHeader("Authorization", "Bearer\t"+d.AccessToken).
		SetHeader("x-share-token", d.ShareToken).
		Post("https://api.aliyundrive.com/v2/file/get_share_link_download_url")
	if err != nil {
		return nil, err
	}
	var u string
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" || e.Code == "ShareLinkTokenInvalid" {
			if e.Code == "AccessTokenInvalid" {
				err = d.refreshToken()
			} else {
				err = d.getShareToken()
			}
			if err != nil {
				return nil, err
			}
			return d.Link(ctx, file, args)
		} else {
			return nil, errors.New(e.Code + ": " + e.Message)
		}
	} else {
		u = resp.DownloadUrl
	}
	return &model.Link{
		Header: http.Header{
			"Referer": []string{"https://www.aliyundrive.com/"},
		},
		URL: u,
	}, nil
}

var _ driver.Driver = (*AliyundriveShare)(nil)
