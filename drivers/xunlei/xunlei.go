package xunlei

import (
	"fmt"
	"sync"
	"time"

	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

var xunleiClient = resty.New().SetTimeout(120 * time.Second)

// 一个账户只允许登陆一次
var userStateCache = struct {
	sync.Mutex
	States map[string]*State
}{States: make(map[string]*State)}

func GetState(account *model.Account) *State {
	userStateCache.Lock()
	defer userStateCache.Unlock()
	if v, ok := userStateCache.States[account.Username]; ok && v != nil {
		return v
	}
	state := new(State).Init()
	userStateCache.States[account.Username] = state
	return state
}

type State struct {
	sync.Mutex
	captchaToken            string
	captchaTokenExpiresTime int64

	tokenType        string
	accessToken      string
	refreshToken     string
	tokenExpiresTime int64 //Milli

	userID string
}

func (s *State) init() *State {
	s.captchaToken = ""
	s.captchaTokenExpiresTime = 0
	s.tokenType = ""
	s.accessToken = ""
	s.refreshToken = ""
	s.tokenExpiresTime = 0
	s.userID = "0"
	return s
}

func (s *State) getToken(account *model.Account) (string, error) {
	if s.isTokensExpires() {
		if err := s.refreshToken_(account); err != nil {
			return "", err
		}
	}
	return fmt.Sprint(s.tokenType, " ", s.accessToken), nil
}

func (s *State) getCaptchaToken(action string, account *model.Account) (string, error) {
	if s.isCaptchaTokenExpires() {
		return s.newCaptchaToken(action, nil, account)
	}
	return s.captchaToken, nil
}

func (s *State) isCaptchaTokenExpires() bool {
	return time.Now().UnixMilli() >= s.captchaTokenExpiresTime || s.captchaToken == "" || s.tokenType == ""
}

func (s *State) isTokensExpires() bool {
	return time.Now().UnixMilli() >= s.tokenExpiresTime || s.accessToken == ""
}

func (s *State) newCaptchaToken(action string, meta map[string]string, account *model.Account) (string, error) {
	ctime := time.Now().UnixMilli()
	driverID := utils.GetMD5Encode(account.Username)
	creq := CaptchaTokenRequest{
		Action:       action,
		CaptchaToken: s.captchaToken,
		ClientID:     CLIENT_ID,
		DeviceID:     driverID,
		Meta: map[string]string{
			"captcha_sign":   captchaSign(driverID, ctime),
			"client_version": CLIENT_VERSION,
			"package_name":   PACKAGE_NAME,
			"timestamp":      fmt.Sprint(ctime),
			"user_id":        s.userID,
		},
	}
	for k, v := range meta {
		creq.Meta[k] = v
	}

	var e Erron
	var resp CaptchaTokenResponse
	_, err := xunleiClient.R().
		SetHeader("X-Device-Id", driverID).
		SetBody(&creq).
		SetError(&e).
		SetResult(&resp).
		Post("https://xluser-ssl.xunlei.com/v1/shield/captcha/init?client_id=" + CLIENT_ID)
	if err != nil {
		return "", err
	}
	if e.ErrorCode != 0 {
		log.Debugf("%+v\n %+v", e, account)
		return "", fmt.Errorf("%s : %s", e.Error, e.ErrorDescription)
	}
	if resp.Url != "" {
		return "", fmt.Errorf("需要验证验证码")
	}

	s.captchaTokenExpiresTime = (ctime + resp.ExpiresIn*1000) - 30000
	s.captchaToken = resp.CaptchaToken
	log.Debugf("%+v\n %+v", s.captchaToken, account)
	return s.captchaToken, nil
}

func (s *State) refreshToken_(account *model.Account) error {
	var e Erron
	var resp TokenResponse
	_, err := xunleiClient.R().
		SetResult(&resp).SetError(&e).
		SetBody(&base.Json{
			"grant_type":    "refresh_token",
			"refresh_token": s.refreshToken,
			"client_id":     CLIENT_ID,
			"client_secret": CLIENT_SECRET,
		}).
		SetHeader("X-Device-Id", utils.GetMD5Encode(account.Username)).SetQueryParam("client_id", CLIENT_ID).
		Post("https://xluser-ssl.xunlei.com/v1/auth/token")
	if err != nil {
		return err
	}

	switch e.ErrorCode {
	case 4122, 4121:
		return s.login(account)
	case 0:
		s.tokenExpiresTime = (time.Now().UnixMilli() + resp.ExpiresIn*1000) - 30000
		s.tokenType = resp.TokenType
		s.accessToken = resp.AccessToken
		s.refreshToken = resp.RefreshToken
		s.userID = resp.UserID
		return nil
	default:
		log.Debugf("%+v\n %+v", e, account)
		return fmt.Errorf("%s : %s", e.Error, e.ErrorDescription)
	}
}

func (s *State) login(account *model.Account) error {
	s.init()
	ctime := time.Now().UnixMilli()
	url := "https://xluser-ssl.xunlei.com/v1/auth/signin"
	captchaToken, err := s.newCaptchaToken(getAction("POST", url), map[string]string{"username": account.Username}, account)
	if err != nil {
		return err
	}

	signReq := SignInRequest{
		CaptchaToken: captchaToken,
		ClientID:     CLIENT_ID,
		ClientSecret: CLIENT_SECRET,
		Username:     account.Username,
		Password:     account.Password,
	}

	var e Erron
	var resp TokenResponse
	_, err = xunleiClient.R().
		SetResult(&resp).
		SetError(&e).
		SetBody(&signReq).
		SetHeader("X-Device-Id", utils.GetMD5Encode(account.Username)).
		SetQueryParam("client_id", CLIENT_ID).
		Post(url)
	if err != nil {
		return err
	}

	defer model.SaveAccount(account)
	if e.ErrorCode != 0 {
		account.Status = e.Error
		log.Debugf("%+v\n %+v", e, account)
		return fmt.Errorf("%s : %s", e.Error, e.ErrorDescription)
	}
	account.Status = "work"
	s.tokenExpiresTime = (ctime + resp.ExpiresIn*1000) - 30000
	s.tokenType = resp.TokenType
	s.accessToken = resp.AccessToken
	s.refreshToken = resp.RefreshToken
	s.userID = resp.UserID
	log.Debugf("%+v\n %+v", resp, account)
	return nil
}

func (s *State) Request(method string, url string, body interface{}, resp interface{}, account *model.Account) error {
	s.Lock()
	token, err := s.getToken(account)
	if err != nil {
		return err
	}

	captchaToken, err := s.getCaptchaToken(getAction(method, url), account)
	if err != nil {
		return err
	}
	s.Unlock()
	var e Erron
	req := xunleiClient.R().
		SetError(&e).
		SetHeader("X-Device-Id", utils.GetMD5Encode(account.Username)).
		SetHeader("Authorization", token).
		SetHeader("X-Captcha-Token", captchaToken).
		SetQueryParam("client_id", CLIENT_ID)

	if body != nil {
		req.SetBody(body)
	}
	if resp != nil {
		req.SetResult(resp)
	}

	switch method {
	case "GET":
		_, err = req.Get(url)
	case "POST":
		_, err = req.Post(url)
	case "DELETE":
		_, err = req.Delete(url)
	case "PATCH":
		_, err = req.Patch(url)
	case "PUT":
		_, err = req.Put(url)
	default:
		return base.ErrNotSupport
	}

	if err != nil {
		return err
	}

	switch e.ErrorCode {
	case 0:
		return nil
	case 9:
		s.newCaptchaToken(getAction(method, url), nil, account)
		fallthrough
	case 4122, 4121:
		return s.Request(method, url, body, resp, account)
	default:
		log.Debugf("%+v\n %+v", e, account)
		return fmt.Errorf("%s : %s", e.Error, e.ErrorDescription)
	}
}

func (s *State) Init() *State {
	s.Lock()
	defer s.Unlock()
	return s.init()
}

func (s *State) GetCaptchaToken(action string, account *model.Account) (string, error) {
	s.Lock()
	defer s.Unlock()
	return s.getCaptchaToken(action, account)
}

func (s *State) GetToken(account *model.Account) (string, error) {
	s.Lock()
	defer s.Unlock()
	return s.getToken(account)
}

func (s *State) Login(account *model.Account) error {
	s.Lock()
	defer s.Unlock()
	return s.login(account)
}
