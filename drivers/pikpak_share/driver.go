package pikpak_share

import (
	"context"
	"net/http"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type PikPakShare struct {
	model.Storage
	Addition
	RefreshToken  string
	AccessToken   string
	PassCodeToken string
}

func (d *PikPakShare) Config() driver.Config {
	return config
}

func (d *PikPakShare) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *PikPakShare) Init(ctx context.Context) error {
	err := d.login()
	if err != nil {
		return err
	}
	if d.SharePwd != "" {
		err = d.getSharePassToken()
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *PikPakShare) Drop(ctx context.Context) error {
	return nil
}

func (d *PikPakShare) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.getFiles(dir.GetID())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return fileToObj(src), nil
	})
}

func (d *PikPakShare) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var resp ShareResp
	query := map[string]string{
		"share_id":        d.ShareId,
		"file_id":         file.GetID(),
		"pass_code_token": d.PassCodeToken,
	}
	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/share/file_info", http.MethodGet, func(req *resty.Request) {
		req.SetQueryParams(query)
	}, &resp)
	if err != nil {
		return nil, err
	}
	link := model.Link{
		URL: resp.FileInfo.WebContentLink,
	}
	if len(resp.FileInfo.Medias) > 0 && resp.FileInfo.Medias[0].Link.Url != "" {
		log.Debugln("use media link")
		link.URL = resp.FileInfo.Medias[0].Link.Url
	}
	return &link, nil
}

func (d *PikPakShare) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	// TODO create folder
	return errs.NotSupport
}

func (d *PikPakShare) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO move obj
	return errs.NotSupport
}

func (d *PikPakShare) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	// TODO rename obj
	return errs.NotSupport
}

func (d *PikPakShare) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj
	return errs.NotSupport
}

func (d *PikPakShare) Remove(ctx context.Context, obj model.Obj) error {
	// TODO remove obj
	return errs.NotSupport
}

func (d *PikPakShare) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// TODO upload file
	return errs.NotSupport
}

var _ driver.Driver = (*PikPakShare)(nil)
