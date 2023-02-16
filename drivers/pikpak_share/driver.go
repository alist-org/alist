package pikpak_share

import (
	"context"
	"net/http"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
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
	return &link, nil
}

var _ driver.Driver = (*PikPakShare)(nil)
