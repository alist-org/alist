package quark_uc_tv

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	"net/http"
	"strconv"
	"time"
)

const (
	UserAgent    = "Mozilla/5.0 (Linux; U; Android 13; zh-cn; M2004J7AC Build/UKQ1.231108.001) AppleWebKit/533.1 (KHTML, like Gecko) Mobile Safari/533.1"
	DeviceBrand  = "Xiaomi"
	Platform     = "tv"
	DeviceName   = "M2004J7AC"
	DeviceModel  = "M2004J7AC"
	BuildDevice  = "M2004J7AC"
	BuildProduct = "M2004J7AC"
	DeviceGpu    = "Adreno (TM) 550"
	ActivityRect = "{}"
)

func (d *QuarkUCTV) request(ctx context.Context, pathname string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	u := d.conf.api + pathname
	tm, token, reqID := d.generateReqSign(method, pathname, d.conf.signKey)
	req := base.RestyClient.R()
	req.SetContext(ctx)
	req.SetHeaders(map[string]string{
		"Accept":          "application/json, text/plain, */*",
		"User-Agent":      UserAgent,
		"x-pan-tm":        tm,
		"x-pan-token":     token,
		"x-pan-client-id": d.conf.clientID,
	})
	req.SetQueryParams(map[string]string{
		"req_id":        reqID,
		"access_token":  d.QuarkUCTVCommon.AccessToken,
		"app_ver":       d.conf.appVer,
		"device_id":     d.Addition.DeviceID,
		"device_brand":  DeviceBrand,
		"platform":      Platform,
		"device_name":   DeviceName,
		"device_model":  DeviceModel,
		"build_device":  BuildDevice,
		"build_product": BuildProduct,
		"device_gpu":    DeviceGpu,
		"activity_rect": ActivityRect,
		"channel":       d.conf.channel,
	})
	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var e Resp
	req.SetError(&e)
	res, err := req.Execute(method, u)
	if err != nil {
		return nil, err
	}
	// 判断 是否需要 刷新 access_token
	if e.Status == -1 && e.Errno == 10001 {
		// token 过期
		err = d.getRefreshTokenByTV(ctx, d.Addition.RefreshToken, true)
		if err != nil {
			return nil, err
		}
		ctx1, cancelFunc := context.WithTimeout(ctx, 10*time.Second)
		defer cancelFunc()
		return d.request(ctx1, pathname, method, callback, resp)
	}

	if e.Status >= 400 || e.Errno != 0 {
		return nil, errors.New(e.ErrorInfo)
	}
	return res.Body(), nil
}

func (d *QuarkUCTV) getLoginCode(ctx context.Context) (string, error) {
	// 获取登录二维码
	pathname := "/oauth/authorize"
	var resp struct {
		CommonRsp
		QrData     string `json:"qr_data"`
		QueryToken string `json:"query_token"`
	}
	_, err := d.request(ctx, pathname, "GET", func(req *resty.Request) {
		req.SetQueryParams(map[string]string{
			"auth_type": "code",
			"client_id": d.conf.clientID,
			"scope":     "netdisk",
			"qrcode":    "1",
			"qr_width":  "460",
			"qr_height": "460",
		})
	}, &resp)
	if err != nil {
		return "", err
	}
	// 保存query_token 用于后续登录
	if resp.QueryToken != "" {
		d.Addition.QueryToken = resp.QueryToken
		op.MustSaveDriverStorage(d)
	}
	return resp.QrData, nil
}

func (d *QuarkUCTV) getCode(ctx context.Context) (string, error) {
	// 通过query token获取code
	pathname := "/oauth/code"
	var resp struct {
		CommonRsp
		Code string `json:"code"`
	}
	_, err := d.request(ctx, pathname, "GET", func(req *resty.Request) {
		req.SetQueryParams(map[string]string{
			"client_id":   d.conf.clientID,
			"scope":       "netdisk",
			"query_token": d.Addition.QueryToken,
		})
	}, &resp)
	if err != nil {
		return "", err
	}
	return resp.Code, nil
}

func (d *QuarkUCTV) getRefreshTokenByTV(ctx context.Context, code string, isRefresh bool) error {
	pathname := "/token"
	_, _, reqID := d.generateReqSign("POST", pathname, d.conf.signKey)
	u := d.conf.codeApi + pathname
	var resp RefreshTokenAuthResp
	body := map[string]string{
		"req_id":        reqID,
		"app_ver":       d.conf.appVer,
		"device_id":     d.Addition.DeviceID,
		"device_brand":  DeviceBrand,
		"platform":      Platform,
		"device_name":   DeviceName,
		"device_model":  DeviceModel,
		"build_device":  BuildDevice,
		"build_product": BuildProduct,
		"device_gpu":    DeviceGpu,
		"activity_rect": ActivityRect,
		"channel":       d.conf.channel,
	}
	if isRefresh {
		body["refresh_token"] = code
	} else {
		body["code"] = code
	}

	_, err := base.RestyClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(&resp).
		SetContext(ctx).
		Post(u)
	if err != nil {
		return err
	}
	if resp.Code != 200 {
		return errors.New(resp.Message)
	}
	if resp.Data.RefreshToken != "" {
		d.Addition.RefreshToken = resp.Data.RefreshToken
		op.MustSaveDriverStorage(d)
		d.QuarkUCTVCommon.AccessToken = resp.Data.AccessToken
	} else {
		return errors.New("refresh token is empty")
	}
	return nil
}

func (d *QuarkUCTV) isLogin(ctx context.Context) (bool, error) {
	_, err := d.request(ctx, "/user", http.MethodGet, func(req *resty.Request) {
		req.SetQueryParams(map[string]string{
			"method": "user_info",
		})
	}, nil)
	return err == nil, err
}

func (d *QuarkUCTV) generateReqSign(method string, pathname string, key string) (string, string, string) {
	//timestamp 13位时间戳
	timestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	deviceID := d.Addition.DeviceID
	if deviceID == "" {
		deviceID = utils.GetMD5EncodeStr(timestamp)
		d.Addition.DeviceID = deviceID
		op.MustSaveDriverStorage(d)
	}
	// 生成req_id
	reqID := md5.Sum([]byte(deviceID + timestamp))
	reqIDHex := hex.EncodeToString(reqID[:])

	// 生成x_pan_token
	tokenData := method + "&" + pathname + "&" + timestamp + "&" + key
	xPanToken := sha256.Sum256([]byte(tokenData))
	xPanTokenHex := hex.EncodeToString(xPanToken[:])

	return timestamp, xPanTokenHex, reqIDHex
}
