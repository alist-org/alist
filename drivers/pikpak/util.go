package pikpak

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/pkg/utils"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/go-resty/resty/v2"
)

// do others that not defined in Driver interface

var Algorithms = []string{
	"PAe56I7WZ6FCSkFy77A96jHWcQA27ui80Qy4",
	"SUbmk67TfdToBAEe2cZyP8vYVeN",
	"1y3yFSZVWiGN95fw/2FQlRuH/Oy6WnO",
	"8amLtHJpGzHPz4m9hGz7r+i+8dqQiAk",
	"tmIEq5yl2g/XWwM3sKZkY4SbL8YUezrvxPksNabUJ",
	"4QvudeJwgJuSf/qb9/wjC21L5aib",
	"D1RJd+FZ+LBbt+dAmaIyYrT9gxJm0BB",
	"1If",
	"iGZr/SJPUFRkwvC174eelKy",
}

const (
	ClientID      = "YNxT9w7GMdWvEOKa"
	ClientSecret  = "dbw2OtmVEeuUvIptb1Coyg"
	ClientVersion = "1.46.2"
	PackageName   = "com.pikcloud.pikpak"
	SdkVersion    = "2.0.4.204000 "
)

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

func (d *PikPak) requestWithCaptchaToken(url string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {

	data, err := d.request(url, method, func(req *resty.Request) {
		req.SetHeaders(map[string]string{
			"User-Agent":      d.GetUserAgent(),
			"X-Device-ID":     d.GetDeviceID(),
			"X-Captcha-Token": d.GetCaptchaToken(),
		})
		if callback != nil {
			callback(req)
		}
	}, resp)

	errResp, ok := err.(*ErrResp)
	if !ok {
		return nil, err
	}

	switch errResp.ErrorCode {
	case 0:
		return data, nil
	//case 4122, 4121, 10, 16:
	//	if d.refreshTokenFunc != nil {
	//		if err = xc.refreshTokenFunc(); err == nil {
	//			break
	//		}
	//	}
	//	return nil, err
	case 9: // 验证码token过期
		if err = d.RefreshCaptchaTokenAtLogin(GetAction(method, url), d.Common.UserID); err != nil {
			return nil, err
		}
	default:
		return nil, err
	}
	return d.request(url, method, callback, resp)
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
	DeviceID  string
	UserAgent string
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
		"client_version": ClientVersion,
		"package_name":   PackageName,
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
	str := fmt.Sprint(ClientID, ClientVersion, PackageName, c.DeviceID, timestamp)
	for _, algorithm := range Algorithms {
		str = utils.GetMD5EncodeStr(str + algorithm)
	}
	sign = "1." + str
	return
}

// 刷新验证码token
func (d *PikPak) refreshCaptchaToken(action string, metas map[string]string) error {
	param := CaptchaTokenRequest{
		Action:       action,
		CaptchaToken: d.Common.CaptchaToken,
		ClientID:     ClientID,
		DeviceID:     d.Common.DeviceID,
		Meta:         metas,
		RedirectUri:  "xlaccsdk01://xbase.cloud/callback?state=harbor",
	}
	var e ErrResp
	var resp CaptchaTokenResponse
	_, err := d.request("https://user.mypikpak.com/v1/shield/captcha/init", http.MethodPost, func(req *resty.Request) {
		req.SetError(&e).SetBody(param)
	}, &resp)

	if err != nil {
		return err
	}

	if e.IsError() {
		return &e
	}

	if resp.Url != "" {
		return fmt.Errorf(`need verify: <a target="_blank" href="%s">Click Here</a>`, resp.Url)
	}

	if resp.CaptchaToken == "" {
		return fmt.Errorf("empty captchaToken")
	}

	if d.Common.RefreshCTokenCk != nil {
		d.Common.RefreshCTokenCk(resp.CaptchaToken)
	}
	d.Common.SetCaptchaToken(resp.CaptchaToken)
	return nil
}
