package template

import (
	"errors"
	"fmt"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
)

// do others that not defined in Driver interface

const (
	DefaultClient = "76lrwrklhdn1icb"
)

func (d *Dropbox) refreshToken() error {
	url := "https://api.dropboxapi.com" + "/oauth/access_token"
	if d.OauthTokenURL != "" && utils.SliceContains([]string{"", "76lrwrklhdn1icb"}, d.ClientID) {
		url = d.OauthTokenURL
	}
	var resp base.TokenResp
	var e struct {
		Error string `json:"error"`
	}
	_, err := base.RestyClient.R().
		ForceContentType("application/json").
		SetBody(base.Json{
			"client_id":     d.ClientID,
			"client_secret": d.ClientSecret,
			"grant_type":    "refresh_token",
			"refresh_token": d.RefreshToken,
		}).
		SetResult(&resp).
		SetError(&e).
		Post(url)
	if err != nil {
		return err
	}
	if e.Error != "" {
		return fmt.Errorf("failed to refresh token: %s", e.Error)
	}
	if resp.RefreshToken == "" {
		return errors.New("failed to refresh token: refresh token is empty")
	}
	d.RefreshToken, d.AccessToken = resp.RefreshToken, resp.AccessToken
	op.MustSaveDriverStorage(d)
	return nil
}
