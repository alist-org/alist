package pikpak

import (
	"errors"
	"net/http"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/go-resty/resty/v2"
)

// do others that not defined in Driver interface

func (d *PikPak) request(url string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R()

	token, err := d.oauth2Token.Token()
	if err != nil {
		return nil, err
	}
	req.SetAuthScheme(token.TokenType).SetAuthToken(token.AccessToken)

	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var e RespErr
	req.SetError(&e)
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}

	if e.ErrorCode != 0 {
		return nil, errors.New(e.Error)
	}
	return res.Body(), nil
}

func (d *PikPak) getFiles(id string) ([]File, error) {
	res := make([]File, 0)
	pageToken := "first"
	for pageToken != "" {
		if pageToken == "first" {
			pageToken = ""
		}
		query := map[string]string{
			"parent_id":      id,
			"thumbnail_size": "SIZE_LARGE",
			"with_audit":     "true",
			"limit":          "100",
			"filters":        `{"phase":{"eq":"PHASE_TYPE_COMPLETE"},"trashed":{"eq":false}}`,
			"page_token":     pageToken,
		}
		var resp Files
		_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files", http.MethodGet, func(req *resty.Request) {
			req.SetQueryParams(query)
		}, &resp)
		if err != nil {
			return nil, err
		}
		pageToken = resp.NextPageToken
		res = append(res, resp.Files...)
	}
	return res, nil
}
