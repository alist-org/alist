package xunlei

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

var xunleiClient = resty.New().
	SetHeaders(map[string]string{
		"Accept": "application/json;charset=UTF-8",
	}).
	SetTimeout(base.DefaultTimeout)

var userClients sync.Map

func GetClient(account *model.Account) *Client {
	if v, ok := userClients.Load(account.Username); ok {
		return v.(*Client)
	}

	client := &Client{
		Client:   xunleiClient,
		driverID: getDriverID(account.Username),
	}
	userClients.Store(account.Username, client)
	return client
}

type Client struct {
	*resty.Client
	sync.Mutex

	driverID     string
	captchaToken string

	token        string
	refreshToken string
	userID       string
}

// 请求验证码token
func (c *Client) requestCaptchaToken(action string, meta map[string]string) error {
	req := CaptchaTokenRequest{
		Action:       action,
		CaptchaToken: c.captchaToken,
		ClientID:     CLIENT_ID,
		DeviceID:     c.driverID,
		Meta:         meta,
	}

	var e Erron
	var resp CaptchaTokenResponse
	_, err := xunleiClient.R().
		SetBody(&req).
		SetError(&e).
		SetResult(&resp).
		Post(XLUSER_API_URL + "/shield/captcha/init")
	if err != nil {
		return err
	}
	if e.ErrorCode != 0 || e.ErrorMsg != "" {
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

	url := XLUSER_API_URL + "/auth/signin"
	err = c.requestCaptchaToken(getAction(http.MethodPost, url), map[string]string{"username": account.Username})
	if err != nil {
		return err
	}

	var e Erron
	var resp TokenResponse
	_, err = xunleiClient.R().
		SetResult(&resp).
		SetError(&e).
		SetBody(&SignInRequest{
			CaptchaToken: c.captchaToken,
			ClientID:     CLIENT_ID,
			ClientSecret: CLIENT_SECRET,
			Username:     account.Username,
			Password:     account.Password,
		}).
		Post(url)
	if err != nil {
		return err
	}

	if e.ErrorCode != 0 || e.ErrorMsg != "" {
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

	ctime := time.Now().UnixMilli()
	return c.requestCaptchaToken(action, map[string]string{
		"captcha_sign":   captchaSign(c.driverID, ctime),
		"client_version": CLIENT_VERSION,
		"package_name":   PACKAGE_NAME,
		"timestamp":      fmt.Sprint(ctime),
		"user_id":        c.userID,
	})
}

// 刷新token
func (c *Client) RefreshToken() error {
	c.Lock()
	defer c.Unlock()

	var e Erron
	var resp TokenResponse
	_, err := xunleiClient.R().
		SetError(&e).
		SetResult(&resp).
		SetBody(&base.Json{
			"grant_type":    "refresh_token",
			"refresh_token": c.refreshToken,
			"client_id":     CLIENT_ID,
			"client_secret": CLIENT_SECRET,
		}).
		Post(XLUSER_API_URL + "/auth/token")
	if err != nil {
		return err
	}
	if e.ErrorCode != 0 || e.ErrorMsg != "" {
		return &e
	}
	c.token = resp.TokenType + " " + resp.AccessToken
	c.refreshToken = resp.RefreshToken
	c.userID = resp.UserID
	return nil
}

func (c *Client) Request(method string, url string, callback func(*resty.Request), account *model.Account) (*resty.Response, error) {
	c.Lock()
	req := xunleiClient.R().
		SetHeaders(map[string]string{
			"X-Device-Id":     c.driverID,
			"Authorization":   c.token,
			"X-Captcha-Token": c.captchaToken,
		}).
		SetQueryParam("client_id", CLIENT_ID)
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
	case 4122, 4121: // token过期
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
