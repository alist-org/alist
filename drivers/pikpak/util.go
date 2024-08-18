package pikpak

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/go-resty/resty/v2"
)

// do others that not defined in Driver interface

var AndroidAlgorithms = []string{
	"Gez0T9ijiI9WCeTsKSg3SMlx",
	"zQdbalsolyb1R/",
	"ftOjr52zt51JD68C3s",
	"yeOBMH0JkbQdEFNNwQ0RI9T3wU/v",
	"BRJrQZiTQ65WtMvwO",
	"je8fqxKPdQVJiy1DM6Bc9Nb1",
	"niV",
	"9hFCW2R1",
	"sHKHpe2i96",
	"p7c5E6AcXQ/IJUuAEC9W6",
	"",
	"aRv9hjc9P+Pbn+u3krN6",
	"BzStcgE8qVdqjEH16l4",
	"SqgeZvL5j9zoHP95xWHt",
	"zVof5yaJkPe3VFpadPof",
}

var WebAlgorithms = []string{
	"C9qPpZLN8ucRTaTiUMWYS9cQvWOE",
	"+r6CQVxjzJV6LCV",
	"F",
	"pFJRC",
	"9WXYIDGrwTCz2OiVlgZa90qpECPD6olt",
	"/750aCr4lm/Sly/c",
	"RB+DT/gZCrbV",
	"",
	"CyLsf7hdkIRxRm215hl",
	"7xHvLi2tOYP0Y92b",
	"ZGTXXxu8E/MIWaEDB+Sm/",
	"1UI3",
	"E7fP5Pfijd+7K+t6Tg/NhuLq0eEUVChpJSkrKxpO",
	"ihtqpG6FMt65+Xk+tWUH2",
	"NhXXU9rg4XXdzo7u5o",
}

const (
	AndroidClientID      = "YNxT9w7GMdWvEOKa"
	AndroidClientSecret  = "dbw2OtmVEeuUvIptb1Coyg"
	AndroidClientVersion = "1.47.1"
	AndroidPackageName   = "com.pikcloud.pikpak"
	AndroidSdkVersion    = "2.0.4.204000"
	WebClientID          = "YUMx5nI8ZU8Ap8pm"
	WebClientSecret      = "dbw2OtmVEeuUvIptb1Coyg"
	WebClientVersion     = "2.0.0"
	WebPackageName       = "mypikpak.com"
	WebSdkVersion        = "8.0.3"
)

func (d *PikPak) login() error {
	url := "https://user.mypikpak.com/v1/auth/signin"
	// 使用 用户填写的 CaptchaToken —————— (验证后的captcha_token)
	if d.GetCaptchaToken() == "" {
		if err := d.RefreshCaptchaTokenInLogin(GetAction(http.MethodPost, url), d.Username); err != nil {
			return err
		}
	}

	var e ErrResp
	res, err := base.RestyClient.SetRetryCount(1).R().SetError(&e).SetBody(base.Json{
		"captcha_token": d.GetCaptchaToken(),
		"client_id":     d.ClientID,
		"client_secret": d.ClientSecret,
		"username":      d.Username,
		"password":      d.Password,
	}).SetQueryParam("client_id", d.ClientID).Post(url)
	if err != nil {
		return err
	}
	if e.ErrorCode != 0 {
		return &e
	}
	data := res.Body()
	d.RefreshToken = jsoniter.Get(data, "refresh_token").ToString()
	d.AccessToken = jsoniter.Get(data, "access_token").ToString()
	d.Common.SetUserID(jsoniter.Get(data, "sub").ToString())
	return nil
}

//func (d *PikPak) refreshToken() error {
//	url := "https://user.mypikpak.com/v1/auth/token"
//	var e ErrResp
//	res, err := base.RestyClient.SetRetryCount(1).R().SetError(&e).
//		SetHeader("user-agent", "").SetBody(base.Json{
//		"client_id":     ClientID,
//		"client_secret": ClientSecret,
//		"grant_type":    "refresh_token",
//		"refresh_token": d.RefreshToken,
//	}).SetQueryParam("client_id", ClientID).Post(url)
//	if err != nil {
//		d.Status = err.Error()
//		op.MustSaveDriverStorage(d)
//		return err
//	}
//	if e.ErrorCode != 0 {
//		if e.ErrorCode == 4126 {
//			// refresh_token invalid, re-login
//			return d.login()
//		}
//		d.Status = e.Error()
//		op.MustSaveDriverStorage(d)
//		return errors.New(e.Error())
//	}
//	data := res.Body()
//	d.Status = "work"
//	d.RefreshToken = jsoniter.Get(data, "refresh_token").ToString()
//	d.AccessToken = jsoniter.Get(data, "access_token").ToString()
//	d.Common.SetUserID(jsoniter.Get(data, "sub").ToString())
//	d.Addition.RefreshToken = d.RefreshToken
//	op.MustSaveDriverStorage(d)
//	return nil
//}

func (d *PikPak) request(url string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		//"Authorization":   "Bearer " + d.AccessToken,
		"User-Agent":      d.GetUserAgent(),
		"X-Device-ID":     d.GetDeviceID(),
		"X-Captcha-Token": d.GetCaptchaToken(),
	})
	if d.oauth2Token != nil {
		// 使用oauth2 获取 access_token
		token, err := d.oauth2Token.Token()
		if err != nil {
			return nil, err
		}
		req.SetAuthScheme(token.TokenType).SetAuthToken(token.AccessToken)
	}

	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var e ErrResp
	req.SetError(&e)
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}

	switch e.ErrorCode {
	case 0:
		return res.Body(), nil
	case 4122, 4121, 16:
		// access_token 过期

		//if err1 := d.refreshToken(); err1 != nil {
		//	return nil, err1
		//}
		t, err := d.oauth2Token.Token()
		if err != nil {
			return nil, err
		}
		d.AccessToken = t.AccessToken
		d.RefreshToken = t.RefreshToken
		d.Addition.RefreshToken = t.RefreshToken
		op.MustSaveDriverStorage(d)

		return d.request(url, method, callback, resp)
	case 9: // 验证码token过期
		if err = d.RefreshCaptchaTokenAtLogin(GetAction(method, url), d.Common.UserID); err != nil {
			return nil, err
		}
		return d.request(url, method, callback, resp)
	case 10: // 操作频繁
		return nil, errors.New(e.ErrorDescription)
	default:
		return nil, errors.New(e.Error())
	}
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

func GetAction(method string, url string) string {
	urlpath := regexp.MustCompile(`://[^/]+((/[^/\s?#]+)*)`).FindStringSubmatch(url)[1]
	return method + ":" + urlpath
}

type Common struct {
	client       *resty.Client
	CaptchaToken string
	UserID       string
	// 必要值,签名相关
	ClientID      string
	ClientSecret  string
	ClientVersion string
	PackageName   string
	Algorithms    []string
	DeviceID      string
	UserAgent     string
	// 验证码token刷新成功回调
	RefreshCTokenCk func(token string)
}

func generateDeviceSign(deviceID, packageName string) string {

	signatureBase := fmt.Sprintf("%s%s%s%s", deviceID, packageName, "1", "appkey")

	sha1Hash := sha1.New()
	sha1Hash.Write([]byte(signatureBase))
	sha1Result := sha1Hash.Sum(nil)

	sha1String := hex.EncodeToString(sha1Result)

	md5Hash := md5.New()
	md5Hash.Write([]byte(sha1String))
	md5Result := md5Hash.Sum(nil)

	md5String := hex.EncodeToString(md5Result)

	deviceSign := fmt.Sprintf("div101.%s%s", deviceID, md5String)

	return deviceSign
}

func BuildCustomUserAgent(deviceID, clientID, appName, sdkVersion, clientVersion, packageName, userID string) string {
	deviceSign := generateDeviceSign(deviceID, packageName)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("ANDROID-%s/%s ", appName, clientVersion))
	sb.WriteString("protocolVersion/200 ")
	sb.WriteString("accesstype/ ")
	sb.WriteString(fmt.Sprintf("clientid/%s ", clientID))
	sb.WriteString(fmt.Sprintf("clientversion/%s ", clientVersion))
	sb.WriteString("action_type/ ")
	sb.WriteString("networktype/WIFI ")
	sb.WriteString("sessionid/ ")
	sb.WriteString(fmt.Sprintf("deviceid/%s ", deviceID))
	sb.WriteString("providername/NONE ")
	sb.WriteString(fmt.Sprintf("devicesign/%s ", deviceSign))
	sb.WriteString("refresh_token/ ")
	sb.WriteString(fmt.Sprintf("sdkversion/%s ", sdkVersion))
	sb.WriteString(fmt.Sprintf("datetime/%d ", time.Now().UnixMilli()))
	sb.WriteString(fmt.Sprintf("usrno/%s ", userID))
	sb.WriteString(fmt.Sprintf("appname/android-%s ", appName))
	sb.WriteString(fmt.Sprintf("session_origin/ "))
	sb.WriteString(fmt.Sprintf("grant_type/ "))
	sb.WriteString(fmt.Sprintf("appid/ "))
	sb.WriteString(fmt.Sprintf("clientip/ "))
	sb.WriteString(fmt.Sprintf("devicename/Xiaomi_M2004j7ac "))
	sb.WriteString(fmt.Sprintf("osversion/13 "))
	sb.WriteString(fmt.Sprintf("platformversion/10 "))
	sb.WriteString(fmt.Sprintf("accessmode/ "))
	sb.WriteString(fmt.Sprintf("devicemodel/M2004J7AC "))

	return sb.String()
}

func (c *Common) SetDeviceID(deviceID string) {
	c.DeviceID = deviceID
}

func (c *Common) SetUserID(userID string) {
	c.UserID = userID
}

func (c *Common) SetUserAgent(userAgent string) {
	c.UserAgent = userAgent
}

func (c *Common) SetCaptchaToken(captchaToken string) {
	c.CaptchaToken = captchaToken
}
func (c *Common) GetCaptchaToken() string {
	return c.CaptchaToken
}

func (c *Common) GetUserAgent() string {
	return c.UserAgent
}

func (c *Common) GetDeviceID() string {
	return c.DeviceID
}

// RefreshCaptchaTokenAtLogin 刷新验证码token(登录后)
func (d *PikPak) RefreshCaptchaTokenAtLogin(action, userID string) error {
	metas := map[string]string{
		"client_version": d.ClientVersion,
		"package_name":   d.PackageName,
		"user_id":        userID,
	}
	metas["timestamp"], metas["captcha_sign"] = d.Common.GetCaptchaSign()
	return d.refreshCaptchaToken(action, metas)
}

// RefreshCaptchaTokenInLogin 刷新验证码token(登录时)
func (d *PikPak) RefreshCaptchaTokenInLogin(action, username string) error {
	metas := make(map[string]string)
	if ok, _ := regexp.MatchString(`\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`, username); ok {
		metas["email"] = username
	} else if len(username) >= 11 && len(username) <= 18 {
		metas["phone_number"] = username
	} else {
		metas["username"] = username
	}
	return d.refreshCaptchaToken(action, metas)
}

// GetCaptchaSign 获取验证码签名
func (c *Common) GetCaptchaSign() (timestamp, sign string) {
	timestamp = fmt.Sprint(time.Now().UnixMilli())
	str := fmt.Sprint(c.ClientID, c.ClientVersion, c.PackageName, c.DeviceID, timestamp)
	for _, algorithm := range c.Algorithms {
		str = utils.GetMD5EncodeStr(str + algorithm)
	}
	sign = "1." + str
	return
}

// refreshCaptchaToken 刷新CaptchaToken
func (d *PikPak) refreshCaptchaToken(action string, metas map[string]string) error {
	param := CaptchaTokenRequest{
		Action:       action,
		CaptchaToken: d.GetCaptchaToken(),
		ClientID:     d.ClientID,
		DeviceID:     d.GetDeviceID(),
		Meta:         metas,
		RedirectUri:  "xlaccsdk01://xbase.cloud/callback?state=harbor",
	}
	var e ErrResp
	var resp CaptchaTokenResponse
	_, err := d.request("https://user.mypikpak.com/v1/shield/captcha/init", http.MethodPost, func(req *resty.Request) {
		req.SetError(&e).SetBody(param).SetQueryParam("client_id", d.ClientID)
	}, &resp)

	if err != nil {
		return err
	}

	if e.IsError() {
		return errors.New(e.Error())
	}

	if resp.Url != "" {
		return fmt.Errorf(`need verify: <a target="_blank" href="%s">Click Here</a>`, resp.Url)
	}

	if d.Common.RefreshCTokenCk != nil {
		d.Common.RefreshCTokenCk(resp.CaptchaToken)
	}
	d.Common.SetCaptchaToken(resp.CaptchaToken)
	return nil
}
