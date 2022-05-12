package xunlei

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

// 缓存登录状态
var userClients sync.Map

func GetClient(account *model.Account) *Client {
	if v, ok := userClients.Load(account.Username); ok {
		return v.(*Client)
	}

	client := &Client{
		Client: base.RestyClient,

		clientID:      account.ClientId,
		clientSecret:  account.ClientSecret,
		clientVersion: account.ClientVersion,
		packageName:   account.PackageName,
		algorithms:    strings.Split(account.Algorithms, ","),
		userAgent:     account.UserAgent,
		deviceID:      account.DeviceId,
	}
	userClients.Store(account.Username, client)
	return client
}

type Client struct {
	*resty.Client
	sync.Mutex

	clientID      string
	clientSecret  string
	clientVersion string
	packageName   string
	algorithms    []string
	userAgent     string
	deviceID      string

	captchaToken string

	token        string
	refreshToken string
	userID       string
}

// 请求验证码token
func (c *Client) requestCaptchaToken(action string, meta map[string]string) error {
	param := CaptchaTokenRequest{
		Action:       action,
		CaptchaToken: c.captchaToken,
		ClientID:     c.clientID,
		DeviceID:     c.deviceID,
		Meta:         meta,
		RedirectUri:  "xlaccsdk01://xunlei.com/callback?state=harbor",
	}

	var e Erron
	var resp CaptchaTokenResponse
	_, err := c.Client.R().
		SetBody(&param).
		SetError(&e).
		SetResult(&resp).
		SetHeader("X-Device-Id", c.deviceID).
		SetQueryParam("client_id", c.clientID).
		Post(XLUSER_API_URL + "/shield/captcha/init")
	if err != nil {
		return err
	}
	if e.HasError() {
		return &e
	}

	if resp.Url != "" {
		return fmt.Errorf("need verify:%s", resp.Url)
	}

	if resp.CaptchaToken == "" {
		return fmt.Errorf("empty captchaToken")
	}
	c.captchaToken = resp.CaptchaToken
	return nil
}

// 验证码签名
func (c *Client) captchaSign(time string) string {
	str := fmt.Sprint(c.clientID, c.clientVersion, c.packageName, c.deviceID, time)
	for _, algorithm := range c.algorithms {
		str = utils.GetMD5Encode(str + algorithm)
	}
	return "1." + str
}

// 登录
func (c *Client) Login(account *model.Account) (err error) {
	c.Lock()
	defer c.Unlock()

	defer func() {
		if err != nil {
			account.Status = err.Error()
		} else {
			account.Status = "work"
		}
		model.SaveAccount(account)
	}()

	meta := make(map[string]string)
	if strings.Contains(account.Username, "@") {
		meta["email"] = account.Username
	} else if len(account.Username) >= 11 {
		if !strings.Contains(account.Username, "+") {
			account.Username = "+86 " + account.Username
		}
		meta["phone_number"] = account.Username
	} else {
		meta["username"] = account.Username
	}

	url := XLUSER_API_URL + "/auth/signin"
	err = c.requestCaptchaToken(getAction(http.MethodPost, url), meta)
	if err != nil {
		return err
	}

	var e Erron
	var resp TokenResponse
	_, err = c.Client.R().
		SetResult(&resp).
		SetError(&e).
		SetBody(&SignInRequest{
			CaptchaToken: c.captchaToken,
			ClientID:     c.clientID,
			ClientSecret: c.clientSecret,
			Username:     account.Username,
			Password:     account.Password,
		}).
		SetHeader("X-Device-Id", c.deviceID).
		SetQueryParam("client_id", c.clientID).
		Post(url)
	if err != nil {
		return err
	}

	if e.HasError() {
		return &e
	}

	if resp.RefreshToken == "" {
		return base.ErrEmptyToken
	}

	c.token = resp.Token()
	c.refreshToken = resp.RefreshToken
	c.userID = resp.UserID
	return nil
}

// 刷新验证码token
func (c *Client) RefreshCaptchaToken(action string) error {
	c.Lock()
	defer c.Unlock()

	timestamp := fmt.Sprint(time.Now().UnixMilli())
	param := map[string]string{
		"client_version": c.clientVersion,
		"package_name":   c.packageName,
		"user_id":        c.userID,
		"captcha_sign":   c.captchaSign(timestamp),
		"timestamp":      timestamp,
	}
	return c.requestCaptchaToken(action, param)
}

// 刷新token
func (c *Client) RefreshToken() error {
	c.Lock()
	defer c.Unlock()

	var e Erron
	var resp TokenResponse
	_, err := c.Client.R().
		SetError(&e).
		SetResult(&resp).
		SetBody(&base.Json{
			"grant_type":    "refresh_token",
			"refresh_token": c.refreshToken,
			"client_id":     c.clientID,
			"client_secret": c.clientSecret,
		}).
		SetHeader("X-Device-Id", c.deviceID).
		SetQueryParam("client_id", c.clientID).
		Post(XLUSER_API_URL + "/auth/token")
	if err != nil {
		return err
	}
	if e.HasError() {
		return &e
	}

	if resp.RefreshToken == "" {
		return base.ErrEmptyToken
	}

	c.token = resp.TokenType + " " + resp.AccessToken
	c.refreshToken = resp.RefreshToken
	c.userID = resp.UserID
	return nil
}

func (c *Client) Request(method string, url string, callback func(*resty.Request), account *model.Account) (*resty.Response, error) {
	c.Lock()
	req := c.Client.R().
		SetHeaders(map[string]string{
			"X-Device-Id":     c.deviceID,
			"Authorization":   c.token,
			"X-Captcha-Token": c.captchaToken,
			"User-Agent":      c.userAgent,
			"client_id":       c.clientID,
		}).
		SetQueryParam("client_id", c.clientID)
	if callback != nil {
		callback(req)
	}
	c.Unlock()

	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}
	log.Debug(res.String())

	var e Erron
	if err = utils.Json.Unmarshal(res.Body(), &e); err != nil {
		return nil, err
	}

	// 处理错误
	switch e.ErrorCode {
	case 0:
		return res, nil
	case 4122, 4121, 10: // token过期
		if err = c.RefreshToken(); err == nil {
			break
		}
		fallthrough
	case 16: // 登录失效
		if err = c.Login(account); err != nil {
			return nil, err
		}
	case 9: // 验证码token过期
		if err = c.RefreshCaptchaToken(getAction(method, url)); err != nil {
			return nil, err
		}
	default:
		return nil, &e
	}
	return c.Request(method, url, callback, account)
}

func (c *Client) UpdateCaptchaToken(captchaToken string) bool {
	c.Lock()
	defer c.Unlock()

	if captchaToken != "" {
		c.captchaToken = captchaToken
		return true
	}
	return false
}
