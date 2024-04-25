package thunder

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

const (
	API_URL        = "https://api-pan.xunlei.com/drive/v1"
	FILE_API_URL   = API_URL + "/files"
	XLUSER_API_URL = "https://xluser-ssl.xunlei.com/v1"
)

const (
	FOLDER    = "drive#folder"
	FILE      = "drive#file"
	RESUMABLE = "drive#resumable"
)

const (
	UPLOAD_TYPE_UNKNOWN = "UPLOAD_TYPE_UNKNOWN"
	//UPLOAD_TYPE_FORM      = "UPLOAD_TYPE_FORM"
	UPLOAD_TYPE_RESUMABLE = "UPLOAD_TYPE_RESUMABLE"
	UPLOAD_TYPE_URL       = "UPLOAD_TYPE_URL"
)

func GetAction(method string, url string) string {
	urlpath := regexp.MustCompile(`://[^/]+((/[^/\s?#]+)*)`).FindStringSubmatch(url)[1]
	return method + ":" + urlpath
}

type Common struct {
	client *resty.Client

	captchaToken string

	// 签名相关,二选一
	Algorithms             []string
	Timestamp, CaptchaSign string

	// 必要值,签名相关
	DeviceID          string
	ClientID          string
	ClientSecret      string
	ClientVersion     string
	PackageName       string
	UserAgent         string
	DownloadUserAgent string
	UseVideoUrl       bool

	// 验证码token刷新成功回调
	refreshCTokenCk func(token string)
}

func (c *Common) SetCaptchaToken(captchaToken string) {
	c.captchaToken = captchaToken
}
func (c *Common) GetCaptchaToken() string {
	return c.captchaToken
}

// 刷新验证码token(登录后)
func (c *Common) RefreshCaptchaTokenAtLogin(action, userID string) error {
	metas := map[string]string{
		"client_version": c.ClientVersion,
		"package_name":   c.PackageName,
		"user_id":        userID,
	}
	metas["timestamp"], metas["captcha_sign"] = c.GetCaptchaSign()
	return c.refreshCaptchaToken(action, metas)
}

// 刷新验证码token(登录时)
func (c *Common) RefreshCaptchaTokenInLogin(action, username string) error {
	metas := make(map[string]string)
	if ok, _ := regexp.MatchString(`\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`, username); ok {
		metas["email"] = username
	} else if len(username) >= 11 && len(username) <= 18 {
		metas["phone_number"] = username
	} else {
		metas["username"] = username
	}
	return c.refreshCaptchaToken(action, metas)
}

// 获取验证码签名
func (c *Common) GetCaptchaSign() (timestamp, sign string) {
	if len(c.Algorithms) == 0 {
		return c.Timestamp, c.CaptchaSign
	}
	timestamp = fmt.Sprint(time.Now().UnixMilli())
	str := fmt.Sprint(c.ClientID, c.ClientVersion, c.PackageName, c.DeviceID, timestamp)
	for _, algorithm := range c.Algorithms {
		str = utils.GetMD5EncodeStr(str + algorithm)
	}
	sign = "1." + str
	return
}

// 刷新验证码token
func (c *Common) refreshCaptchaToken(action string, metas map[string]string) error {
	param := CaptchaTokenRequest{
		Action:       action,
		CaptchaToken: c.captchaToken,
		ClientID:     c.ClientID,
		DeviceID:     c.DeviceID,
		Meta:         metas,
		RedirectUri:  "xlaccsdk01://xunlei.com/callback?state=harbor",
	}
	var e ErrResp
	var resp CaptchaTokenResponse
	_, err := c.Request(XLUSER_API_URL+"/shield/captcha/init", http.MethodPost, func(req *resty.Request) {
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

	if c.refreshCTokenCk != nil {
		c.refreshCTokenCk(resp.CaptchaToken)
	}
	c.SetCaptchaToken(resp.CaptchaToken)
	return nil
}

// 只有基础信息的请求
func (c *Common) Request(url, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := c.client.R().SetHeaders(map[string]string{
		"user-agent":       c.UserAgent,
		"accept":           "application/json;charset=UTF-8",
		"x-device-id":      c.DeviceID,
		"x-client-id":      c.ClientID,
		"x-client-version": c.ClientVersion,
	})

	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}

	var erron ErrResp
	utils.Json.Unmarshal(res.Body(), &erron)
	if erron.IsError() {
		return nil, &erron
	}

	return res.Body(), nil
}

// 计算文件Gcid
func getGcid(r io.Reader, size int64) (string, error) {
	calcBlockSize := func(j int64) int64 {
		var psize int64 = 0x40000
		for float64(j)/float64(psize) > 0x200 && psize < 0x200000 {
			psize = psize << 1
		}
		return psize
	}

	hash1 := sha1.New()
	hash2 := sha1.New()
	readSize := calcBlockSize(size)
	for {
		hash2.Reset()
		if n, err := utils.CopyWithBufferN(hash2, r, readSize); err != nil && n == 0 {
			if err != io.EOF {
				return "", err
			}
			break
		}
		hash1.Write(hash2.Sum(nil))
	}
	return hex.EncodeToString(hash1.Sum(nil)), nil
}
