package _189pc

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/utils"

	"github.com/avast/retry-go"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
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

func (y *Cloud189PC) SignatureHeader(url, method, params string) map[string]string {
	dateOfGmt := getHttpDateStr()
	sessionKey := y.tokenInfo.SessionKey
	sessionSecret := y.tokenInfo.SessionSecret
	if y.isFamily() {
		sessionKey = y.tokenInfo.FamilySessionKey
		sessionSecret = y.tokenInfo.FamilySessionSecret
	}

	header := map[string]string{
		"Date":         dateOfGmt,
		"SessionKey":   sessionKey,
		"X-Request-ID": uuid.NewString(),
		"Signature":    signatureOfHmac(sessionSecret, sessionKey, method, url, dateOfGmt, params),
	}
	return header
}

func (y *Cloud189PC) EncryptParams(params Params) string {
	sessionSecret := y.tokenInfo.SessionSecret
	if y.isFamily() {
		sessionSecret = y.tokenInfo.FamilySessionSecret
	}
	if params != nil {
		return AesECBEncrypt(params.Encode(), sessionSecret[:16])
	}
	return ""
}

func (y *Cloud189PC) request(url, method string, callback base.ReqCallback, params Params, resp interface{}) ([]byte, error) {
	req := y.client.R().SetQueryParams(clientSuffix())

	// 设置params
	paramsData := y.EncryptParams(params)
	if paramsData != "" {
		req.SetQueryParam("params", paramsData)
	}

	// Signature
	req.SetHeaders(y.SignatureHeader(url, method, paramsData))

	var erron RespErr
	req.SetError(&erron)

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

	if strings.Contains(res.String(), "userSessionBO is null") {
		if err = y.refreshSession(); err != nil {
			return nil, err
		}
		return y.request(url, method, callback, params, resp)
	}

	// 处理错误
	if erron.HasError() {
		if erron.ErrorCode == "InvalidSessionKey" {
			if err = y.refreshSession(); err != nil {
				return nil, err
			}
			return y.request(url, method, callback, params, resp)
		}
		return nil, &erron
	}
	return res.Body(), nil
}

func (y *Cloud189PC) get(url string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	return y.request(url, http.MethodGet, callback, nil, resp)
}

func (y *Cloud189PC) post(url string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	return y.request(url, http.MethodPost, callback, nil, resp)
}

func (y *Cloud189PC) put(ctx context.Context, url string, headers map[string]string, sign bool, file io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, file)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	for key, value := range clientSuffix() {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	if sign {
		for key, value := range y.SignatureHeader(url, http.MethodPut, "") {
			req.Header.Add(key, value)
		}
	}

	resp, err := base.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var erron RespErr
	jsoniter.Unmarshal(body, &erron)
	xml.Unmarshal(body, &erron)
	if erron.HasError() {
		return nil, &erron
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("put fail,err:%s", string(body))
	}
	return body, nil
}
func (y *Cloud189PC) getFiles(ctx context.Context, fileId string) ([]model.Obj, error) {
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

func (y *Cloud189PC) login() (err error) {
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
		// 遇到错误，重新加载登陆参数(刷新验证码)
		if err != nil && y.NoUseOcr {
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
			// "mailSuffix":   "@189.cn",
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

	if erron.HasError() {
		return &erron
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
func (y *Cloud189PC) initLoginParam() error {
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
	y.loginParam = &param

	// 判断是否需要验证码
	resp, err := y.client.R().
		SetHeader("REQID", param.ReqId).
		SetFormData(map[string]string{
			"appKey":      APP_ID,
			"accountType": ACCOUNT_TYPE,
			"userName":    param.RsaUsername,
		}).Post(AUTH_URL + "/api/logbox/oauth2/needcaptcha.do")
	if err != nil {
		return err
	}
	if resp.String() == "0" {
		return nil
	}

	// 拉取验证码
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
	if imgRes.Size() > 20 {
		if setting.GetStr(conf.OcrApi) != "" && !y.NoUseOcr {
			vRes, err := base.RestyClient.R().
				SetMultipartField("image", "validateCode.png", "image/png", bytes.NewReader(imgRes.Body())).
				Post(setting.GetStr(conf.OcrApi))
			if err != nil {
				return err
			}
			if jsoniter.Get(vRes.Body(), "status").ToInt() == 200 {
				y.VCode = jsoniter.Get(vRes.Body(), "result").ToString()
				return nil
			}
		}

		// 返回验证码图片给前端
		return fmt.Errorf(`need img validate code: <img src="data:image/png;base64,%s"/>`, base64.StdEncoding.EncodeToString(imgRes.Body()))
	}
	return nil
}

// 刷新会话
func (y *Cloud189PC) refreshSession() (err error) {
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

	if erron.HasError() {
		if erron.ResCode == "UserInvalidOpenToken" {
			if err = y.login(); err != nil {
				return err
			}
		}
		return &erron
	}
	y.tokenInfo.UserSessionResp = userSessionResp
	return
}

// 普通上传
// 无法上传大小为0的文件
func (y *Cloud189PC) StreamUpload(ctx context.Context, dstDir model.Obj, file model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	var DEFAULT = partSize(file.GetSize())
	var count = int(math.Ceil(float64(file.GetSize()) / float64(DEFAULT)))

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
	_, err := y.request(fullUrl+"/initMultiUpload", http.MethodGet, func(req *resty.Request) {
		req.SetContext(ctx)
	}, params, &initMultiUpload)
	if err != nil {
		return nil, err
	}

	fileMd5 := md5.New()
	silceMd5 := md5.New()
	silceMd5Hexs := make([]string, 0, count)
	byteData := bytes.NewBuffer(make([]byte, DEFAULT))
	for i := 1; i <= count; i++ {
		if utils.IsCanceled(ctx) {
			return nil, ctx.Err()
		}

		// 读取块
		byteData.Reset()
		silceMd5.Reset()
		_, err := io.CopyN(io.MultiWriter(fileMd5, silceMd5, byteData), file, DEFAULT)
		if err != io.EOF && err != io.ErrUnexpectedEOF && err != nil {
			return nil, err
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
			return nil, err
		}

		// 开始上传
		uploadData := uploadUrl.UploadUrls[fmt.Sprint("partNumber_", i)]

		err = retry.Do(func() error {
			_, err := y.put(ctx, uploadData.RequestURL, ParseHttpHeader(uploadData.RequestHeader), false, bytes.NewReader(byteData.Bytes()))
			return err
		},
			retry.Context(ctx),
			retry.Attempts(3),
			retry.Delay(time.Second),
			retry.MaxDelay(5*time.Second))
		if err != nil {
			return nil, err
		}
		up(int(i * 100 / count))
	}

	fileMd5Hex := strings.ToUpper(hex.EncodeToString(fileMd5.Sum(nil)))
	sliceMd5Hex := fileMd5Hex
	if file.GetSize() > DEFAULT {
		sliceMd5Hex = strings.ToUpper(utils.GetMD5EncodeStr(strings.Join(silceMd5Hexs, "\n")))
	}

	// 提交上传
	var resp CommitMultiUploadFileResp
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
		}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.toFile(), nil
}

// 快传
func (y *Cloud189PC) FastUpload(ctx context.Context, dstDir model.Obj, file model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	// 需要获取完整文件md5,必须支持 io.Seek
	tempFile, err := utils.CreateTempFile(file.GetReadCloser())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	var DEFAULT = partSize(file.GetSize())
	count := int(math.Ceil(float64(file.GetSize()) / float64(DEFAULT)))

	// 优先计算所需信息
	fileMd5 := md5.New()
	silceMd5 := md5.New()
	silceMd5Hexs := make([]string, 0, count)
	silceMd5Base64s := make([]string, 0, count)
	for i := 1; i <= count; i++ {
		if utils.IsCanceled(ctx) {
			return nil, ctx.Err()
		}

		silceMd5.Reset()
		if _, err := io.CopyN(io.MultiWriter(fileMd5, silceMd5), tempFile, DEFAULT); err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return nil, err
		}
		md5Byte := silceMd5.Sum(nil)
		silceMd5Hexs = append(silceMd5Hexs, strings.ToUpper(hex.EncodeToString(md5Byte)))
		silceMd5Base64s = append(silceMd5Base64s, fmt.Sprint(i, "-", base64.StdEncoding.EncodeToString(md5Byte)))
	}
	if _, err = tempFile.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	fileMd5Hex := strings.ToUpper(hex.EncodeToString(fileMd5.Sum(nil)))
	sliceMd5Hex := fileMd5Hex
	if file.GetSize() > DEFAULT {
		sliceMd5Hex = strings.ToUpper(utils.GetMD5EncodeStr(strings.Join(silceMd5Hexs, "\n")))
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
		return nil, err
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
			return nil, err
		}

		buf := make([]byte, DEFAULT)
		for i := 1; i <= count; i++ {
			if utils.IsCanceled(ctx) {
				return nil, ctx.Err()
			}

			n, err := io.ReadFull(tempFile, buf)
			if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				return nil, err
			}
			uploadData := uploadUrls.UploadUrls[fmt.Sprint("partNumber_", i)]
			err = retry.Do(func() error {
				_, err := y.put(ctx, uploadData.RequestURL, ParseHttpHeader(uploadData.RequestHeader), false, bytes.NewReader(buf[:n]))
				return err
			},
				retry.Context(ctx),
				retry.Attempts(3),
				retry.Delay(time.Second),
				retry.MaxDelay(5*time.Second))
			if err != nil {
				return nil, err
			}

			up(int(i * 100 / count))
		}
	}

	// 提交
	var resp CommitMultiUploadFileResp
	_, err = y.request(fullUrl+"/commitMultiUploadFile", http.MethodGet,
		func(req *resty.Request) {
			req.SetContext(ctx)
		}, Params{
			"uploadFileId": uploadInfo.Data.UploadFileID,
			"isLog":        "0",
			"opertype":     "3",
		}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.toFile(), nil
}

// 旧版本上传，家庭云不支持覆盖
func (y *Cloud189PC) OldUpload(ctx context.Context, dstDir model.Obj, file model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	// 需要获取完整文件md5,必须支持 io.Seek
	tempFile, err := utils.CreateTempFile(file.GetReadCloser())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	// 计算md5
	fileMd5 := md5.New()
	if _, err := io.Copy(fileMd5, tempFile); err != nil {
		return nil, err
	}
	if _, err = tempFile.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	fileMd5Hex := strings.ToUpper(hex.EncodeToString(fileMd5.Sum(nil)))

	// 创建上传会话
	var uploadInfo CreateUploadFileResp

	fullUrl := API_URL + "/createUploadFile.action"
	if y.isFamily() {
		fullUrl = API_URL + "/family/file/createFamilyFile.action"
	}
	_, err = y.post(fullUrl, func(req *resty.Request) {
		req.SetContext(ctx)
		if y.isFamily() {
			req.SetQueryParams(map[string]string{
				"familyId":     y.FamilyID,
				"fileMd5":      fileMd5Hex,
				"fileName":     file.GetName(),
				"fileSize":     fmt.Sprint(file.GetSize()),
				"parentId":     dstDir.GetID(),
				"resumePolicy": "1",
			})
		} else {
			req.SetFormData(map[string]string{
				"parentFolderId": dstDir.GetID(),
				"fileName":       file.GetName(),
				"size":           fmt.Sprint(file.GetSize()),
				"md5":            fileMd5Hex,
				"opertype":       "3",
				"flag":           "1",
				"resumePolicy":   "1",
				"isLog":          "0",
				// "baseFileId":     "",
				// "lastWrite":"",
				// "localPath": strings.ReplaceAll(param.LocalPath, "\\", "/"),
				// "fileExt": "",
			})
		}
	}, &uploadInfo)

	if err != nil {
		return nil, err
	}

	// 网盘中不存在该文件，开始上传
	status := GetUploadFileStatusResp{CreateUploadFileResp: uploadInfo}
	for status.Size < file.GetSize() && status.FileDataExists != 1 {
		if utils.IsCanceled(ctx) {
			return nil, ctx.Err()
		}

		header := map[string]string{
			"ResumePolicy": "1",
			"Expect":       "100-continue",
		}

		if y.isFamily() {
			header["FamilyId"] = fmt.Sprint(y.FamilyID)
			header["UploadFileId"] = fmt.Sprint(status.UploadFileId)
		} else {
			header["Edrive-UploadFileId"] = fmt.Sprint(status.UploadFileId)
		}

		_, err := y.put(ctx, status.FileUploadUrl, header, true, io.NopCloser(tempFile))
		if err, ok := err.(*RespErr); ok && err.Code != "InputStreamReadError" {
			return nil, err
		}

		// 获取断点状态
		fullUrl := API_URL + "/getUploadFileStatus.action"
		if y.isFamily() {
			fullUrl = API_URL + "/family/file/getFamilyFileStatus.action"
		}
		_, err = y.get(fullUrl, func(req *resty.Request) {
			req.SetContext(ctx).SetQueryParams(map[string]string{
				"uploadFileId": fmt.Sprint(status.UploadFileId),
				"resumePolicy": "1",
			})
			if y.isFamily() {
				req.SetQueryParam("familyId", fmt.Sprint(y.FamilyID))
			}
		}, &status)
		if err != nil {
			return nil, err
		}

		if _, err := tempFile.Seek(status.GetSize(), io.SeekStart); err != nil {
			return nil, err
		}
		up(int(status.Size / file.GetSize()))
	}

	// 提交
	var resp OldCommitUploadFileResp
	_, err = y.post(status.FileCommitUrl, func(req *resty.Request) {
		req.SetContext(ctx)
		if y.isFamily() {
			req.SetHeaders(map[string]string{
				"ResumePolicy": "1",
				"UploadFileId": fmt.Sprint(status.UploadFileId),
				"FamilyId":     fmt.Sprint(y.FamilyID),
			})
		} else {
			req.SetFormData(map[string]string{
				"opertype":     "3",
				"resumePolicy": "1",
				"uploadFileId": fmt.Sprint(status.UploadFileId),
				"isLog":        "0",
			})
		}
	}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.toFile(), nil
}

func (y *Cloud189PC) isFamily() bool {
	return y.Type == "family"
}

func (y *Cloud189PC) isLogin() bool {
	if y.tokenInfo == nil {
		return false
	}
	_, err := y.get(API_URL+"/getUserInfo.action", nil, nil)
	return err == nil
}

// 获取家庭云所有用户信息
func (y *Cloud189PC) getFamilyInfoList() ([]FamilyInfoResp, error) {
	var resp FamilyInfoListResp
	_, err := y.get(API_URL+"/family/manage/getFamilyList.action", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.FamilyInfoResp, nil
}

// 抽取家庭云ID
func (y *Cloud189PC) getFamilyID() (string, error) {
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

func (y *Cloud189PC) CheckBatchTask(aType string, taskID string) (*BatchTaskStateResp, error) {
	var resp BatchTaskStateResp
	_, err := y.post(API_URL+"/batch/checkBatchTask.action", func(req *resty.Request) {
		req.SetFormData(map[string]string{
			"type":   aType,
			"taskId": taskID,
		})
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (y *Cloud189PC) WaitBatchTask(aType string, taskID string, t time.Duration) error {
	for {
		state, err := y.CheckBatchTask(aType, taskID)
		if err != nil {
			return err
		}
		switch state.TaskStatus {
		case 2:
			return errors.New("there is a conflict with the target object")
		case 4:
			return nil
		}
		time.Sleep(t)
	}
}
