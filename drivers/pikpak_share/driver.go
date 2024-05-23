package pikpak_share

import (
	"context"
	"net/http"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	"golang.org/x/oauth2"
)

type PikPakShare struct {
	model.Storage
	Addition
	oauth2Token   oauth2.TokenSource
	PassCodeToken string
}

func (d *PikPakShare) Config() driver.Config {
	return config
}

func (d *PikPakShare) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *PikPakShare) Init(ctx context.Context) error {
	if d.ClientID == "" || d.ClientSecret == "" {
		d.ClientID = "YNxT9w7GMdWvEOKa"
		d.ClientSecret = "dbw2OtmVEeuUvIptb1Coyg"
	}

	oauth2Config := &oauth2.Config{
		ClientID:     d.ClientID,
		ClientSecret: d.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://user.mypikpak.com/v1/auth/signin",
			TokenURL:  "https://user.mypikpak.com/v1/auth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	d.oauth2Token = oauth2.ReuseTokenSource(nil, utils.TokenSource(func() (*oauth2.Token, error) {
		return oauth2Config.PasswordCredentialsToken(
			context.WithValue(context.Background(), oauth2.HTTPClient, base.HttpClient),
			d.Username,
			d.Password,
		)
	}))

	if d.SharePwd != "" {
		return d.getSharePassToken()
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

	downloadUrl := resp.FileInfo.WebContentLink
	if downloadUrl == "" && len(resp.FileInfo.Medias) > 0 {
		downloadUrl = resp.FileInfo.Medias[0].Link.Url
	}

	link := model.Link{
		URL: downloadUrl,
	}
	return &link, nil
}

var _ driver.Driver = (*PikPakShare)(nil)
