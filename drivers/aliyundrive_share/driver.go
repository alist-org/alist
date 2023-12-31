package aliyundrive_share

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Xhofe/rateg"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/cron"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type AliyundriveShare struct {
	model.Storage
	Addition
	AccessToken string
	ShareToken  string
	DriveId     string
	cron        *cron.Cron

	limitList func(ctx context.Context, dir model.Obj) ([]model.Obj, error)
	limitLink func(ctx context.Context, file model.Obj) (*model.Link, error)
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

func (d *AliyundriveShare) Drop(ctx context.Context) error {
	if d.cron != nil {
		d.cron.Stop()
	}
	d.DriveId = ""
	return nil
}

func (d *AliyundriveShare) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	if d.limitList == nil {
		return nil, fmt.Errorf("driver not init")
	}
	return d.limitList(ctx, dir)
}

func (d *AliyundriveShare) list(ctx context.Context, dir model.Obj) ([]model.Obj, error) {
	files, err := d.getFiles(dir.GetID())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return fileToObj(src), nil
	})
}

func (d *AliyundriveShare) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	if d.limitLink == nil {
		return nil, fmt.Errorf("driver not init")
	}
	return d.limitLink(ctx, file)
}

func (d *AliyundriveShare) link(ctx context.Context, file model.Obj) (*model.Link, error) {
	data := base.Json{
		"drive_id": d.DriveId,
		"file_id":  file.GetID(),
		// // Only ten minutes lifetime
		"expire_sec": 600,
		"share_id":   d.ShareId,
	}
	var resp ShareLinkResp
	_, err := d.request("https://api.alipan.com/v2/file/get_share_link_download_url", http.MethodPost, func(req *resty.Request) {
		req.SetHeader(CanaryHeaderKey, CanaryHeaderValue).SetBody(data).SetResult(&resp)
	})
	if err != nil {
		return nil, err
	}
	return &model.Link{
		Header: http.Header{
			"Referer": []string{"https://www.alipan.com/"},
		},
		URL: resp.DownloadUrl,
	}, nil
}

func (d *AliyundriveShare) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
	var resp base.Json
	var url string
	data := base.Json{
		"share_id": d.ShareId,
		"file_id":  args.Obj.GetID(),
	}
	switch args.Method {
	case "doc_preview":
		url = "https://api.alipan.com/v2/file/get_office_preview_url"
	case "video_preview":
		url = "https://api.alipan.com/v2/file/get_video_preview_play_info"
		data["category"] = "live_transcoding"
	default:
		return nil, errs.NotSupport
	}
	_, err := d.request(url, http.MethodPost, func(req *resty.Request) {
		req.SetBody(data).SetResult(&resp)
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

var _ driver.Driver = (*AliyundriveShare)(nil)
