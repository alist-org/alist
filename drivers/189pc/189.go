package _189

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"sync"

	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
)

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
	state := &State{client: resty.New().
		SetHeaders(map[string]string{
			"Accept":     "application/json;charset=UTF-8",
			"User-Agent": base.UserAgent,
		}),
	}
	userStateCache.States[account.Username] = state
	return state
}

type State struct {
	sync.Mutex
	client *resty.Client

	RsaPublicKey string

	SessionKey          string
	SessionSecret       string
	FamilySessionKey    string
	FamilySessionSecret string

	AccessToken string

	//怎么刷新的???
	RefreshToken string
}

func (s *State) login(account *model.Account) error {
	// 清除cookie
	jar, _ := cookiejar.New(nil)
	s.client.SetCookieJar(jar)

	var err error
	var res *resty.Response
	defer func() {
		account.Status = "work"
		if err != nil {
			account.Status = err.Error()
		}
		model.SaveAccount(account)
		if res != nil {
			log.Debug(res.String())
		}
	}()

	var param *LoginParam
	param, err = s.getLoginParam()
	if err != nil {
		return err
	}

	// 提交登录
	s.RsaPublicKey = fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", param.jRsaKey)
	res, err = s.client.R().
		SetHeaders(map[string]string{
			"Referer": AUTH_URL,
			"REQID":   param.ReqId,
			"lt":      param.Lt,
		}).
		SetFormData(map[string]string{
			"appKey":       APP_ID,
			"accountType":  "02",
			"userName":     "{RSA}" + rsaEncrypt(s.RsaPublicKey, account.Username),
			"password":     "{RSA}" + rsaEncrypt(s.RsaPublicKey, account.Password),
			"validateCode": param.vCodeRS,
			"captchaToken": param.CaptchaToken,
			"returnUrl":    RETURN_URL,
			"mailSuffix":   "@189.cn",
			"dynamicCheck": "FALSE",
			"clientType":   CLIENT_TYPE,
			"cb_SaveName":  "1",
			"isOauth2":     "false",
			"state":        "",
			"paramId":      param.ParamId,
		}).
		Post(AUTH_URL + "/api/logbox/oauth2/loginSubmit.do")
	if err != nil {
		return err
	}
	toUrl := jsoniter.Get(res.Body(), "toUrl").ToString()
	if toUrl == "" {
		log.Error(res.String())
		return fmt.Errorf(res.String())
	}

	// 获取Session
	var erron Erron
	var sessionResp appSessionResp
	res, err = s.client.R().
		SetResult(&sessionResp).SetError(&erron).
		SetQueryParams(clientSuffix()).
		SetQueryParam("redirectURL", url.QueryEscape(toUrl)).
		Post(API_URL + "/getSessionForPC.action")
	if err != nil {
		return err
	}

	if erron.ResCode != "" {
		err = fmt.Errorf(erron.ResMessage)
		return err
	}
	if sessionResp.ResCode != 0 {
		err = fmt.Errorf(sessionResp.ResMessage)
		return err
	}
	s.SessionKey = sessionResp.SessionKey
	s.SessionSecret = sessionResp.SessionSecret
	s.FamilySessionKey = sessionResp.FamilySessionKey
	s.FamilySessionSecret = sessionResp.FamilySessionSecret
	s.AccessToken = sessionResp.AccessToken
	s.RefreshToken = sessionResp.RefreshToken
	return err
}

func (s *State) getLoginParam() (*LoginParam, error) {
	res, err := s.client.R().
		SetQueryParams(map[string]string{
			"appId":      APP_ID,
			"clientType": CLIENT_TYPE,
			"returnURL":  RETURN_URL,
			"timeStamp":  fmt.Sprint(timestamp()),
		}).
		Get(WEB_URL + "/api/portal/unifyLoginForPC.action")
	if err != nil {
		return nil, err
	}
	log.Debug(res.String())
	param := &LoginParam{
		CaptchaToken: regexp.MustCompile(`'captchaToken' value='(.+?)'`).FindStringSubmatch(res.String())[1],
		Lt:           regexp.MustCompile(`lt = "(.+?)"`).FindStringSubmatch(res.String())[1],
		ParamId:      regexp.MustCompile(`paramId = "(.+?)"`).FindStringSubmatch(res.String())[1],
		ReqId:        regexp.MustCompile(`reqId = "(.+?)"`).FindStringSubmatch(res.String())[1],
		jRsaKey:      regexp.MustCompile(`"j_rsaKey" value="(.+?)"`).FindStringSubmatch(res.String())[1],

		vCodeID: regexp.MustCompile(`token=([A-Za-z0-9&=]+)`).FindStringSubmatch(res.String())[1],
	}

	imgRes, err := s.client.R().Get(fmt.Sprint(AUTH_URL, "/api/logbox/oauth2/picCaptcha.do?token=", param.vCodeID, timestamp()))
	if err != nil {
		return nil, err
	}
	if len(imgRes.Body()) > 0 {
		vRes, err := resty.New().R().
			SetMultipartField("image", "validateCode.png", "image/png", bytes.NewReader(imgRes.Body())).
			Post(conf.GetStr("ocr api"))
		if err != nil {
			return nil, err
		}
		if jsoniter.Get(vRes.Body(), "status").ToInt() != 200 {
			return nil, errors.New("ocr error:" + jsoniter.Get(vRes.Body(), "msg").ToString())
		}
		param.vCodeRS = jsoniter.Get(vRes.Body(), "result").ToString()
		log.Debugln("code: ", param.vCodeRS)
	}
	return param, nil
}

func (s *State) refreshSession(account *model.Account) error {
	var erron Erron
	var userSessionResp UserSessionResp
	res, err := s.client.R().
		SetResult(&userSessionResp).SetError(&erron).
		SetQueryParams(clientSuffix()).
		SetQueryParams(map[string]string{
			"appId":       APP_ID,
			"accessToken": s.AccessToken,
		}).
		SetHeader("X-Request-ID", uuid.NewString()).
		Get("https://api.cloud.189.cn/getSessionForPC.action")
	if err != nil {
		return err
	}
	log.Debug(res.String())
	if erron.ResCode != "" {
		return fmt.Errorf(erron.ResMessage)
	}

	switch userSessionResp.ResCode {
	case 0:
		s.SessionKey = userSessionResp.SessionKey
		s.SessionSecret = userSessionResp.SessionSecret
		s.FamilySessionKey = userSessionResp.FamilySessionKey
		s.FamilySessionSecret = userSessionResp.FamilySessionSecret
	case 11, 18:
		return s.login(account)
	default:
		account.Status = userSessionResp.ResMessage
		_ = model.SaveAccount(account)
		return fmt.Errorf(userSessionResp.ResMessage)
	}
	return nil
}

func (s *State) IsLogin() bool {
	_, err := s.Request("GET", API_URL+"/getUserInfo.action", nil, func(r *resty.Request) {
		r.SetQueryParams(clientSuffix())
	}, nil)
	return err == nil
}

func (s *State) Login(account *model.Account) error {
	s.Lock()
	defer s.Unlock()
	return s.login(account)
}

func (s *State) RefreshSession(account *model.Account) error {
	s.Lock()
	defer s.Unlock()
	return s.refreshSession(account)
}

func (s *State) Request(method string, fullUrl string, params url.Values, callback func(*resty.Request), account *model.Account) (*resty.Response, error) {
	s.Lock()
	dateOfGmt := getHttpDateStr()
	sessionKey := s.SessionKey
	sessionSecret := s.SessionSecret
	if account != nil && isFamily(account) {
		sessionKey = s.FamilySessionKey
		sessionSecret = s.FamilySessionSecret
	}

	req := s.client.R()
	req.SetHeaders(map[string]string{
		"Date":         dateOfGmt,
		"SessionKey":   sessionKey,
		"X-Request-ID": uuid.NewString(),
	})

	// 设置params
	var paramsData string
	if params != nil {
		paramsData = AesECBEncrypt(params.Encode(), s.SessionSecret[:16])
		req.SetQueryParam("params", paramsData)
	}
	req.SetHeader("Signature", signatureOfHmac(sessionSecret, sessionKey, method, fullUrl, dateOfGmt, paramsData))

	callback(req)
	s.Unlock()

	var err error
	var res *resty.Response
	switch method {
	case "GET":
		res, err = req.Get(fullUrl)
	case "POST":
		res, err = req.Post(fullUrl)
	case "DELETE":
		res, err = req.Delete(fullUrl)
	case "PATCH":
		res, err = req.Patch(fullUrl)
	case "PUT":
		res, err = req.Put(fullUrl)
	default:
		return nil, base.ErrNotSupport
	}
	if err != nil {
		return nil, err
	}
	log.Debug(res.String())

	var erron Erron
	utils.Json.Unmarshal(res.Body(), &erron)
	if erron.ResCode != "" {
		return nil, fmt.Errorf(erron.ResMessage)
	}
	if erron.Code != "" && erron.Code != "SUCCESS" {
		if erron.Msg == "" {
			return nil, fmt.Errorf(erron.Message)
		}
		return nil, fmt.Errorf(erron.Msg)
	}
	if erron.ErrorCode != "" {
		return nil, fmt.Errorf(erron.ErrorMsg)
	}

	if account != nil {
		switch jsoniter.Get(res.Body(), "res_code").ToInt64() {
		case 11, 18:
			if err := s.refreshSession(account); err != nil {
				return nil, err
			}
			return s.Request(method, fullUrl, params, callback, account)
		case 0:
			if res.StatusCode() == http.StatusOK {
				return res, nil
			}
			fallthrough
		default:
			return nil, fmt.Errorf(res.String())
		}
	}

	if jsoniter.Get(res.Body(), "res_code").ToInt64() != 0 {
		return res, fmt.Errorf(jsoniter.Get(res.Body(), "res_message").ToString())
	}
	return res, nil
}
