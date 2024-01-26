package aliyundrive

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/dustinxie/ecc"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

func (d *AliDrive) createSession() error {
	state, ok := global.Load(d.UserID)
	if !ok {
		return fmt.Errorf("can't load user state, user_id: %s", d.UserID)
	}
	d.sign()
	state.retry++
	if state.retry > 3 {
		state.retry = 0
		return fmt.Errorf("createSession failed after three retries")
	}
	_, err, _ := d.request("https://api.alipan.com/users/v1/users/device/create_session", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"deviceName":   "samsung",
			"modelName":    "SM-G9810",
			"nonce":        0,
			"pubKey":       PublicKeyToHex(&state.privateKey.PublicKey),
			"refreshToken": d.RefreshToken,
		})
	}, nil)
	if err == nil{
		state.retry = 0
	}
	return err
}

// func (d *AliDrive) renewSession() error {
// 	_, err, _ := d.request("https://api.alipan.com/users/v1/users/device/renew_session", http.MethodPost, nil, nil)
// 	return err
// }

func (d *AliDrive) sign() {
	state, _ := global.Load(d.UserID)
	secpAppID := "5dde4e1bdf9e4966b387ba58f4b3fdc3"
	singdata := fmt.Sprintf("%s:%s:%s:%d", secpAppID, state.deviceID, d.UserID, 0)
	hash := sha256.Sum256([]byte(singdata))
	data, _ := ecc.SignBytes(state.privateKey, hash[:], ecc.RecID|ecc.LowerS)
	state.signature = hex.EncodeToString(data) //strconv.Itoa(state.nonce)
}

// do others that not defined in Driver interface

func (d *AliDrive) refreshToken() error {
	url := "https://auth.alipan.com/v2/account/token"
	var resp base.TokenResp
	var e RespErr
	_, err := base.RestyClient.R().
		//ForceContentType("application/json").
		SetBody(base.Json{"refresh_token": d.RefreshToken, "grant_type": "refresh_token"}).
		SetResult(&resp).
		SetError(&e).
		Post(url)
	if err != nil {
		return err
	}
	if e.Code != "" {
		return fmt.Errorf("failed to refresh token: %s", e.Message)
	}
	if resp.RefreshToken == "" {
		return errors.New("failed to refresh token: refresh token is empty")
	}
	d.RefreshToken, d.AccessToken = resp.RefreshToken, resp.AccessToken
	op.MustSaveDriverStorage(d)
	return nil
}

func (d *AliDrive) request(url, method string, callback base.ReqCallback, resp interface{}) ([]byte, error, RespErr) {
	req := base.RestyClient.R()
	state, ok := global.Load(d.UserID)
	if !ok {
		if url == "https://api.alipan.com/v2/user/get" {
			state = &State{}
		} else {
			return nil, fmt.Errorf("can't load user state, user_id: %s", d.UserID), RespErr{}
		}
	}
	req.SetHeaders(map[string]string{
		"Authorization": "Bearer\t" + d.AccessToken,
		"content-type":  "application/json",
		"origin":        "https://www.alipan.com",
		"Referer":       "https://alipan.com/",
		"X-Signature":   state.signature,
		"x-request-id":  uuid.NewString(),
		"X-Canary":      "client=Android,app=adrive,version=v4.1.0",
		"X-Device-Id":   state.deviceID,
	})
	if callback != nil {
		callback(req)
	} else {
		req.SetBody("{}")
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var e RespErr
	req.SetError(&e)
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err, e
	}
	if e.Code != "" {
		switch e.Code {
		case "AccessTokenInvalid":
			err = d.refreshToken()
			if err != nil {
				return nil, err, e
			}
		case "DeviceSessionSignatureInvalid":
			err = d.createSession()
			if err != nil {
				return nil, err, e
			}
		default:
			return nil, errors.New(e.Message), e
		}
		return d.request(url, method, callback, resp)
	} else if res.IsError() {
		return nil, errors.New("bad status code " + res.Status()), e
	}
	return res.Body(), nil, e
}

func (d *AliDrive) getFiles(fileId string) ([]File, error) {
	marker := "first"
	res := make([]File, 0)
	for marker != "" {
		if marker == "first" {
			marker = ""
		}
		var resp Files
		data := base.Json{
			"drive_id":                d.DriveId,
			"fields":                  "*",
			"image_thumbnail_process": "image/resize,w_400/format,jpeg",
			"image_url_process":       "image/resize,w_1920/format,jpeg",
			"limit":                   200,
			"marker":                  marker,
			"order_by":                d.OrderBy,
			"order_direction":         d.OrderDirection,
			"parent_file_id":          fileId,
			"video_thumbnail_process": "video/snapshot,t_0,f_jpg,ar_auto,w_300",
			"url_expire_sec":          14400,
		}
		_, err, _ := d.request("https://api.alipan.com/v2/file/list", http.MethodPost, func(req *resty.Request) {
			req.SetBody(data)
		}, &resp)

		if err != nil {
			return nil, err
		}
		marker = resp.NextMarker
		res = append(res, resp.Items...)
	}
	return res, nil
}

func (d *AliDrive) batch(srcId, dstId string, url string) error {
	res, err, _ := d.request("https://api.alipan.com/v3/batch", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"requests": []base.Json{
				{
					"headers": base.Json{
						"Content-Type": "application/json",
					},
					"method": "POST",
					"id":     srcId,
					"body": base.Json{
						"drive_id":          d.DriveId,
						"file_id":           srcId,
						"to_drive_id":       d.DriveId,
						"to_parent_file_id": dstId,
					},
					"url": url,
				},
			},
			"resource": "file",
		})
	}, nil)
	if err != nil {
		return err
	}
	status := utils.Json.Get(res, "responses", 0, "status").ToInt()
	if status < 400 && status >= 100 {
		return nil
	}
	return errors.New(string(res))
}
