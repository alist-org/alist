package dropbox

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

func (d *Dropbox) refreshToken() error {
	url := d.base + "/oauth2/token"
	if utils.SliceContains([]string{"", DefaultClientID}, d.ClientID) {
		url = d.OauthTokenURL
	}
	var tokenResp TokenResp
	resp, err := base.RestyClient.R().
		//ForceContentType("application/x-www-form-urlencoded").
		//SetBasicAuth(d.ClientID, d.ClientSecret).
		SetFormData(map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": d.RefreshToken,
			"client_id":     d.ClientID,
			"client_secret": d.ClientSecret,
		}).
		Post(url)
	if err != nil {
		return err
	}
	log.Debugf("[dropbox] refresh token response: %s", resp.String())
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to refresh token: %s", resp.String())
	}
	_ = utils.Json.UnmarshalFromString(resp.String(), &tokenResp)
	d.AccessToken = tokenResp.AccessToken
	op.MustSaveDriverStorage(d)
	return nil
}

func (d *Dropbox) request(uri, method string, callback base.ReqCallback, retry ...bool) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeader("Authorization", "Bearer "+d.AccessToken)
	if d.RootNamespaceId != "" {
		apiPathRootJson, err := utils.Json.MarshalToString(map[string]interface{}{
			".tag": "root",
			"root": d.RootNamespaceId,
		})
		if err != nil {
			return nil, err
		}
		req.SetHeader("Dropbox-API-Path-Root", apiPathRootJson)
	}
	if callback != nil {
		callback(req)
	}
	if method == http.MethodPost && req.Body != nil {
		req.SetHeader("Content-Type", "application/json")
	}
	var e ErrorResp
	req.SetError(&e)
	res, err := req.Execute(method, d.base+uri)
	if err != nil {
		return nil, err
	}
	log.Debugf("[dropbox] request (%s) response: %s", uri, res.String())
	isRetry := len(retry) > 0 && retry[0]
	if res.StatusCode() != 200 {
		body := res.String()
		if !isRetry && (utils.SliceMeet([]string{"expired_access_token", "invalid_access_token", "authorization"}, body,
			func(item string, v string) bool {
				return strings.Contains(v, item)
			}) || d.AccessToken == "") {
			err = d.refreshToken()
			if err != nil {
				return nil, err
			}
			return d.request(uri, method, callback, true)
		}
		return nil, fmt.Errorf("%s:%s", e.Error, e.ErrorSummary)
	}
	return res.Body(), nil
}

func (d *Dropbox) list(ctx context.Context, data base.Json, isContinue bool) (*ListResp, error) {
	var resp ListResp
	uri := "/2/files/list_folder"
	if isContinue {
		uri += "/continue"
	}
	_, err := d.request(uri, http.MethodPost, func(req *resty.Request) {
		req.SetContext(ctx).SetBody(data).SetResult(&resp)
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (d *Dropbox) getFiles(ctx context.Context, path string) ([]File, error) {
	hasMore := true
	var marker string
	res := make([]File, 0)

	data := base.Json{
		"include_deleted":                     false,
		"include_has_explicit_shared_members": false,
		"include_mounted_folders":             false,
		"include_non_downloadable_files":      false,
		"limit":                               2000,
		"path":                                path,
		"recursive":                           false,
	}
	resp, err := d.list(ctx, data, false)
	if err != nil {
		return nil, err
	}
	marker = resp.Cursor
	hasMore = resp.HasMore
	res = append(res, resp.Entries...)

	for hasMore {
		data := base.Json{
			"cursor": marker,
		}
		resp, err := d.list(ctx, data, true)
		if err != nil {
			return nil, err
		}
		marker = resp.Cursor
		hasMore = resp.HasMore
		res = append(res, resp.Entries...)
	}
	return res, nil
}

func (d *Dropbox) finishUploadSession(ctx context.Context, toPath string, offset int64, sessionId string) error {
	url := d.contentBase + "/2/files/upload_session/finish"
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer "+d.AccessToken)

	uploadFinishArgs := UploadFinishArgs{
		Commit: struct {
			Autorename     bool   `json:"autorename"`
			Mode           string `json:"mode"`
			Mute           bool   `json:"mute"`
			Path           string `json:"path"`
			StrictConflict bool   `json:"strict_conflict"`
		}{
			Autorename:     true,
			Mode:           "add",
			Mute:           false,
			Path:           toPath,
			StrictConflict: false,
		},
		Cursor: UploadCursor{
			Offset:    offset,
			SessionID: sessionId,
		},
	}

	argsJson, err := utils.Json.MarshalToString(uploadFinishArgs)
	if err != nil {
		return err
	}
	req.Header.Set("Dropbox-API-Arg", argsJson)

	res, err := base.HttpClient.Do(req)
	if err != nil {
		log.Errorf("failed to update file when finish session, err: %+v", err)
		return err
	}
	_ = res.Body.Close()
	return nil
}

func (d *Dropbox) startUploadSession(ctx context.Context) (string, error) {
	url := d.contentBase + "/2/files/upload_session/start"
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer "+d.AccessToken)
	req.Header.Set("Dropbox-API-Arg", "{\"close\":false}")

	res, err := base.HttpClient.Do(req)
	if err != nil {
		log.Errorf("failed to update file when start session, err: %+v", err)
		return "", err
	}

	body, err := io.ReadAll(res.Body)
	sessionId := utils.Json.Get(body, "session_id").ToString()

	_ = res.Body.Close()
	return sessionId, nil
}
