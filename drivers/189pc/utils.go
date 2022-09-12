package _189pc

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
)

const (
	ACCOUNT_TYPE = "02"
	APP_ID       = "8025431004"
	CLIENT_TYPE  = "10020"
	VERSION      = "6.2"

	WEB_URL    = "https://cloud.189.cn"
	AUTH_URL   = "https://open.e.189.cn"
	API_URL    = "https://api.cloud.189.cn"
	UPLOAD_URL = "https://upload.cloud.189.cn"

	RETURN_URL = "https://m.cloud.189.cn/zhuanti/2020/loginErrorPc/index.html"

	PC  = "TELEPC"
	MAC = "TELEMAC"

	CHANNEL_ID = "web_cloud.189.cn"
)

func (y *Yun189PC) request(url, method string, callback base.ReqCallback, params Params, resp interface{}) ([]byte, error) {
	dateOfGmt := getHttpDateStr()
	sessionKey := y.tokenInfo.SessionKey
	sessionSecret := y.tokenInfo.SessionSecret
	if y.isFamily() {
		sessionKey = y.tokenInfo.FamilySessionKey
		sessionSecret = y.tokenInfo.FamilySessionSecret
	}

	req := y.client.R().SetQueryParams(clientSuffix()).SetHeaders(map[string]string{
		"Date":         dateOfGmt,
		"SessionKey":   sessionKey,
		"X-Request-ID": uuid.NewString(),
	})

	// 设置params
	var paramsData string
	if params != nil {
		paramsData = AesECBEncrypt(params.Encode(), sessionSecret[:16])
		req.SetQueryParam("params", paramsData)
	}
	req.SetHeader("Signature", signatureOfHmac(sessionSecret, sessionKey, method, url, dateOfGmt, paramsData))

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
	var erron RespErr
	utils.Json.Unmarshal(res.Body(), &erron)

	if erron.ResCode != "" {
		return nil, fmt.Errorf("res_code: %s ,res_msg: %s", erron.ResCode, erron.ResMessage)
	}
	if erron.Code != "" && erron.Code != "SUCCESS" {
		if erron.Msg != "" {
			return nil, fmt.Errorf("code: %s ,msg: %s", erron.Code, erron.Msg)
		}
		if erron.Message != "" {
			return nil, fmt.Errorf("code: %s ,msg: %s", erron.Code, erron.Message)
		}
		return nil, fmt.Errorf(res.String())
	}
	switch erron.ErrorCode {
	case "":
		break
	case "InvalidSessionKey":
		if err = y.refreshSession(); err != nil {
			return nil, err
		}
		return y.request(url, method, callback, params, resp)
	default:
		return nil, fmt.Errorf("err_code: %s ,err_msg: %s", erron.ErrorCode, erron.ErrorMsg)
	}

	if strings.Contains(res.String(), "userSessionBO is null") {
		if err = y.refreshSession(); err != nil {
			return nil, err
		}
		return y.request(url, method, callback, params, resp)
	}

	resCode := utils.Json.Get(res.Body(), "res_code").ToInt64()
	message := utils.Json.Get(res.Body(), "res_message").ToString()
	switch resCode {
	case 0:
		return res.Body(), nil
	default:
		return nil, fmt.Errorf("res_code: %d ,res_msg: %s", resCode, message)
	}
}

func (y *Yun189PC) get(url string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	return y.request(url, http.MethodGet, callback, nil, resp)
}

func (y *Yun189PC) post(url string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	return y.request(url, http.MethodPost, callback, nil, resp)
}

func (y *Yun189PC) getFiles(ctx context.Context, fileId string) ([]model.Obj, error) {
	fullUrl := API_URL
	if y.isFamily() {
		fullUrl += "/family/file"
	}
	fullUrl += "/listFiles.action"

	res := make([]model.Obj, 0, 130)
	for pageNum := 1; ; pageNum++ {
		var resp Cloud189FilesResp
		_, err := y.get(fullUrl, func(r *resty.Request) {
			r.SetContext(ctx)
			r.SetQueryParams(map[string]string{
				"folderId":   fileId,
				"fileType":   "0",
				"mediaAttr":  "0",
				"iconOption": "5",
				"pageNum":    fmt.Sprint(pageNum),
				"pageSize":   "130",
			})
			if y.isFamily() {
				r.SetQueryParams(map[string]string{
					"familyId":   y.FamilyID,
					"orderBy":    toFamilyOrderBy(y.OrderBy),
					"descending": toDesc(y.OrderDirection),
				})
			} else {
				r.SetQueryParams(map[string]string{
					"recursive":  "0",
					"orderBy":    y.OrderBy,
					"descending": toDesc(y.OrderDirection),
				})
			}
		}, &resp)
		if err != nil {
			return nil, err
		}
		// 获取完毕跳出
		if resp.FileListAO.Count == 0 {
			break
		}

		for i := 0; i < len(resp.FileListAO.FolderList); i++ {
			res = append(res, &resp.FileListAO.FolderList[i])
		}
		for i := 0; i < len(resp.FileListAO.FileList); i++ {
			res = append(res, &resp.FileListAO.FileList[i])
		}
	}
	return res, nil
}

func (y *Yun189PC) login() (err error) {
	// 初始化登陆所需参数
	if y.loginParam == nil {
		if err = y.initLoginParam(); err != nil {
			// 验证码也通过错误返回
			return err
		}
	}

	defer func() {
		// 销毁验证码
		y.VCode = ""
		// 销毁登陆参数
		y.loginParam = nil
		// 遇到错误，重新加载登陆参数
		if err != nil {
			if err1 := y.initLoginParam(); err1 != nil {
				err = fmt.Errorf("err1: %s \nerr2: %s", err, err1)
			}
		}
	}()

	param := y.loginParam
	var loginresp LoginResp
	_, err = y.client.R().
		ForceContentType("application/json;charset=UTF-8").SetResult(&loginresp).
		SetHeaders(map[string]string{
			"REQID": param.ReqId,
			"lt":    param.Lt,
		}).
		SetFormData(map[string]string{
			"appKey":       APP_ID,
			"accountType":  ACCOUNT_TYPE,
			"userName":     param.RsaUsername,
			"password":     param.RsaPassword,
			"validateCode": y.VCode,
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
	if loginresp.ToUrl == "" {
		return fmt.Errorf("login failed,No toUrl obtained, msg: %s", loginresp.Msg)
	}

	// 获取Session
	var erron RespErr
	var tokenInfo AppSessionResp
	_, err = y.client.R().
		SetResult(&tokenInfo).SetError(&erron).
		SetQueryParams(clientSuffix()).
		SetQueryParam("redirectURL", url.QueryEscape(loginresp.ToUrl)).
		Post(API_URL + "/getSessionForPC.action")
	if err != nil {
		return
	}

	if erron.ResCode != "" {
		err = fmt.Errorf(erron.ResMessage)
		return
	}
	if tokenInfo.ResCode != 0 {
		err = fmt.Errorf(tokenInfo.ResMessage)
		return
	}
	y.tokenInfo = &tokenInfo
	return
}

/* 初始化登陆需要的参数
*  如果遇到验证码返回错误
 */
func (y *Yun189PC) initLoginParam() error {
	// 清除cookie
	jar, _ := cookiejar.New(nil)
	y.client.SetCookieJar(jar)

	res, err := y.client.R().
		SetQueryParams(map[string]string{
			"appId":      APP_ID,
			"clientType": CLIENT_TYPE,
			"returnURL":  RETURN_URL,
			"timeStamp":  fmt.Sprint(timestamp()),
		}).
		Get(WEB_URL + "/api/portal/unifyLoginForPC.action")
	if err != nil {
		return err
	}

	param := LoginParam{
		CaptchaToken: regexp.MustCompile(`'captchaToken' value='(.+?)'`).FindStringSubmatch(res.String())[1],
		Lt:           regexp.MustCompile(`lt = "(.+?)"`).FindStringSubmatch(res.String())[1],
		ParamId:      regexp.MustCompile(`paramId = "(.+?)"`).FindStringSubmatch(res.String())[1],
		ReqId:        regexp.MustCompile(`reqId = "(.+?)"`).FindStringSubmatch(res.String())[1],
		// jRsaKey:      regexp.MustCompile(`"j_rsaKey" value="(.+?)"`).FindStringSubmatch(res.String())[1],
	}

	// 获取rsa公钥
	var encryptConf EncryptConfResp
	_, err = y.client.R().
		ForceContentType("application/json;charset=UTF-8").SetResult(&encryptConf).
		SetFormData(map[string]string{"appId": APP_ID}).
		Post(AUTH_URL + "/api/logbox/config/encryptConf.do")
	if err != nil {
		return err
	}

	param.jRsaKey = fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", encryptConf.Data.PubKey)
	param.RsaUsername = encryptConf.Data.Pre + RsaEncrypt(param.jRsaKey, y.Username)
	param.RsaPassword = encryptConf.Data.Pre + RsaEncrypt(param.jRsaKey, y.Password)

	// 判断是否需要验证码
	res, err = y.client.R().
		SetFormData(map[string]string{
			"appKey":      APP_ID,
			"accountType": ACCOUNT_TYPE,
			"userName":    param.RsaUsername,
		}).
		Post(AUTH_URL + "/api/logbox/oauth2/needcaptcha.do")
	if err != nil {
		return err
	}

	y.loginParam = &param
	if res.String() != "0" {
		imgRes, err := y.client.R().
			SetQueryParams(map[string]string{
				"token": param.CaptchaToken,
				"REQID": param.ReqId,
				"rnd":   fmt.Sprint(timestamp()),
			}).
			Get(AUTH_URL + "/api/logbox/oauth2/picCaptcha.do")
		if err != nil {
			return fmt.Errorf("failed to obtain verification code")
		}

		// 尝试使用ocr
		vRes, err := base.RestyClient.R().
			SetMultipartField("image", "validateCode.png", "image/png", bytes.NewReader(imgRes.Body())).
			Post(setting.GetStr(conf.OcrApi))
		if err == nil && jsoniter.Get(vRes.Body(), "status").ToInt() == 200 {
			y.VCode = jsoniter.Get(vRes.Body(), "result").ToString()
		}

		// ocr无法处理，返回验证码图片给前端
		if len(y.VCode) != 4 {
			return fmt.Errorf("need validate code: data:image/png;base64,%s", base64.StdEncoding.EncodeToString(res.Body()))
		}
	}
	return nil
}

// 刷新会话
func (y *Yun189PC) refreshSession() (err error) {
	var erron RespErr
	var userSessionResp UserSessionResp
	_, err = y.client.R().
		SetResult(&userSessionResp).SetError(&erron).
		SetQueryParams(clientSuffix()).
		SetQueryParams(map[string]string{
			"appId":       APP_ID,
			"accessToken": y.tokenInfo.AccessToken,
		}).
		SetHeader("X-Request-ID", uuid.NewString()).
		Get(API_URL + "/getSessionForPC.action")
	if err != nil {
		return err
	}

	// 错误影响正常访问，下线该储存
	defer func() {
		if err != nil {
			y.GetStorage().SetStatus(fmt.Sprintf("%+v", err.Error()))
			op.MustSaveDriverStorage(y)
		}
	}()

	switch erron.ResCode {
	case "":
		break
	case "UserInvalidOpenToken":
		if err = y.login(); err != nil {
			return err
		}
	default:
		err = fmt.Errorf("res_code: %s ,res_msg: %s", erron.ResCode, erron.ResMessage)
		return
	}

	switch userSessionResp.ResCode {
	case 0:
		y.tokenInfo.UserSessionResp = userSessionResp
	default:
		err = fmt.Errorf("code: %d , msg: %s", userSessionResp.ResCode, userSessionResp.ResMessage)
	}
	return
}

// 普通上传
func (y *Yun189PC) CommonUpload(ctx context.Context, dstDir model.Obj, file model.FileStreamer, up driver.UpdateProgress) (err error) {
	const DEFAULT int64 = 10485760
	var count = int64(math.Ceil(float64(file.GetSize()) / float64(DEFAULT)))

	params := Params{
		"parentFolderId": dstDir.GetID(),
		"fileName":       url.QueryEscape(file.GetName()),
		"fileSize":       fmt.Sprint(file.GetSize()),
		"sliceSize":      fmt.Sprint(DEFAULT),
		"lazyCheck":      "1",
	}

	fullUrl := UPLOAD_URL
	if y.isFamily() {
		params.Set("familyId", y.FamilyID)
		fullUrl += "/family"
	} else {
		//params.Set("extend", `{"opScene":"1","relativepath":"","rootfolderid":""}`)
		fullUrl += "/person"
	}

	// 初始化上传
	var initMultiUpload InitMultiUploadResp
	_, err = y.request(fullUrl+"/initMultiUpload", http.MethodGet, func(req *resty.Request) {
		req.SetContext(ctx)
	}, params, &initMultiUpload)
	if err != nil {
		return err
	}

	fileMd5 := md5.New()
	silceMd5 := md5.New()
	silceMd5Hexs := make([]string, 0, count)
	byteData := bytes.NewBuffer(make([]byte, DEFAULT))
	for i := int64(1); i <= count; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 读取块
		byteData.Reset()
		silceMd5.Reset()
		_, err := io.CopyN(io.MultiWriter(fileMd5, silceMd5, byteData), file, DEFAULT)
		if err != io.EOF && err != io.ErrUnexpectedEOF && err != nil {
			return err
		}

		// 计算块md5并进行hex和base64编码
		md5Bytes := silceMd5.Sum(nil)
		silceMd5Hexs = append(silceMd5Hexs, strings.ToUpper(hex.EncodeToString(md5Bytes)))
		silceMd5Base64 := base64.StdEncoding.EncodeToString(md5Bytes)

		// 获取上传链接
		var uploadUrl UploadUrlsResp
		_, err = y.request(fullUrl+"/getMultiUploadUrls", http.MethodGet,
			func(req *resty.Request) {
				req.SetContext(ctx)
			}, Params{
				"partInfo":     fmt.Sprintf("%d-%s", i, silceMd5Base64),
				"uploadFileId": initMultiUpload.Data.UploadFileID,
			}, &uploadUrl)
		if err != nil {
			return err
		}

		// 开始上传
		uploadData := uploadUrl.UploadUrls[fmt.Sprint("partNumber_", i)]
		res, err := y.putClient.R().
			SetContext(ctx).
			SetQueryParams(clientSuffix()).
			SetHeaders(ParseHttpHeader(uploadData.RequestHeader)).
			SetBody(byteData).
			Put(uploadData.RequestURL)
		if err != nil {
			return err
		}
		if res.StatusCode() != http.StatusOK {
			return fmt.Errorf("updload fail,msg: %s", res.String())
		}
		up(int(i * 100 / count))
	}

	fileMd5Hex := strings.ToUpper(hex.EncodeToString(fileMd5.Sum(nil)))
	sliceMd5Hex := fileMd5Hex
	if file.GetSize() > DEFAULT {
		sliceMd5Hex = strings.ToUpper(utils.GetMD5Encode(strings.Join(silceMd5Hexs, "\n")))
	}

	// 提交上传
	_, err = y.request(fullUrl+"/commitMultiUploadFile", http.MethodGet,
		func(req *resty.Request) {
			req.SetContext(ctx)
		}, Params{
			"uploadFileId": initMultiUpload.Data.UploadFileID,
			"fileMd5":      fileMd5Hex,
			"sliceMd5":     sliceMd5Hex,
			"lazyCheck":    "1",
			"isLog":        "0",
			"opertype":     "3",
		}, nil)
	return err
}

// 快传
func (y *Yun189PC) FastUpload(ctx context.Context, dstDir model.Obj, file model.FileStreamer, up driver.UpdateProgress) (err error) {
	// 需要获取完整文件md5,必须支持 io.Seek
	tempFile, err := utils.CreateTempFile(file.GetReadCloser())
	if err != nil {
		return err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	const DEFAULT int64 = 10485760
	count := int(math.Ceil(float64(file.GetSize()) / float64(DEFAULT)))

	// 优先计算所需信息
	fileMd5 := md5.New()
	silceMd5 := md5.New()
	silceMd5Hexs := make([]string, 0, count)
	silceMd5Base64s := make([]string, 0, count)
	for i := 1; i <= count; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		silceMd5.Reset()
		if _, err := io.CopyN(io.MultiWriter(fileMd5, silceMd5), tempFile, DEFAULT); err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err
		}
		md5Byte := silceMd5.Sum(nil)
		silceMd5Hexs = append(silceMd5Hexs, strings.ToUpper(hex.EncodeToString(md5Byte)))
		silceMd5Base64s = append(silceMd5Base64s, fmt.Sprint(i, "-", base64.StdEncoding.EncodeToString(md5Byte)))
	}
	if _, err = tempFile.Seek(0, io.SeekStart); err != nil {
		return err
	}

	fileMd5Hex := strings.ToUpper(hex.EncodeToString(fileMd5.Sum(nil)))
	sliceMd5Hex := fileMd5Hex
	if file.GetSize() > DEFAULT {
		sliceMd5Hex = strings.ToUpper(utils.GetMD5Encode(strings.Join(silceMd5Hexs, "\n")))
	}

	// 检测是否支持快传
	params := Params{
		"parentFolderId": dstDir.GetID(),
		"fileName":       url.QueryEscape(file.GetName()),
		"fileSize":       fmt.Sprint(file.GetSize()),
		"fileMd5":        fileMd5Hex,
		"sliceSize":      fmt.Sprint(DEFAULT),
		"sliceMd5":       sliceMd5Hex,
	}

	fullUrl := UPLOAD_URL
	if y.isFamily() {
		params.Set("familyId", y.FamilyID)
		fullUrl += "/family"
	} else {
		//params.Set("extend", `{"opScene":"1","relativepath":"","rootfolderid":""}`)
		fullUrl += "/person"
	}

	var uploadInfo InitMultiUploadResp
	_, err = y.request(fullUrl+"/initMultiUpload", http.MethodGet, func(req *resty.Request) {
		req.SetContext(ctx)
	}, params, &uploadInfo)
	if err != nil {
		return err
	}

	// 网盘中不存在该文件，开始上传
	if uploadInfo.Data.FileDataExists != 1 {
		var uploadUrls UploadUrlsResp
		_, err = y.request(fullUrl+"/getMultiUploadUrls", http.MethodGet,
			func(req *resty.Request) {
				req.SetContext(ctx)
			}, Params{
				"uploadFileId": uploadInfo.Data.UploadFileID,
				"partInfo":     strings.Join(silceMd5Base64s, ","),
			}, &uploadUrls)
		if err != nil {
			return err
		}

		for i := 1; i <= count; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			uploadData := uploadUrls.UploadUrls[fmt.Sprint("partNumber_", i)]
			res, err := y.putClient.R().
				SetContext(ctx).
				SetQueryParams(clientSuffix()).
				SetHeaders(ParseHttpHeader(uploadData.RequestHeader)).
				SetBody(io.LimitReader(tempFile, DEFAULT)).
				Put(uploadData.RequestURL)
			if err != nil {
				return err
			}
			if res.StatusCode() != http.StatusOK {
				return fmt.Errorf("updload fail,msg: %s", res.String())
			}
			up(int(i * 100 / count))
		}
	}

	// 提交
	_, err = y.request(fullUrl+"/commitMultiUploadFile", http.MethodGet,
		func(req *resty.Request) {
			req.SetContext(ctx)
		}, Params{
			"uploadFileId": uploadInfo.Data.UploadFileID,
			"isLog":        "0",
			"opertype":     "3",
		}, nil)
	return err
}

func (y *Yun189PC) isFamily() bool {
	return y.Type == "family"
}

func (y *Yun189PC) isLogin() bool {
	if y.tokenInfo == nil {
		return false
	}
	_, err := y.get(API_URL+"/getUserInfo.action", nil, nil)
	return err == nil
}

// 获取家庭云所有用户信息
func (y *Yun189PC) getFamilyInfoList() ([]FamilyInfoResp, error) {
	var resp FamilyInfoListResp
	_, err := y.get(API_URL+"/family/manage/getFamilyList.action", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.FamilyInfoResp, nil
}

// 抽取家庭云ID
func (y *Yun189PC) getFamilyID() (string, error) {
	infos, err := y.getFamilyInfoList()
	if err != nil {
		return "", err
	}
	if len(infos) == 0 {
		return "", fmt.Errorf("cannot get automatically,please input family_id")
	}
	for _, info := range infos {
		if strings.Contains(y.tokenInfo.LoginName, info.RemarkName) {
			return fmt.Sprint(info.FamilyID), nil
		}
	}
	return fmt.Sprint(infos[0].FamilyID), nil
}
