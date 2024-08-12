package _189pc

import (
	"bytes"
	"container/ring"
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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/errgroup"
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

func (y *Cloud189PC) SignatureHeader(url, method, params string, isFamily bool) map[string]string {
	dateOfGmt := getHttpDateStr()
	sessionKey := y.tokenInfo.SessionKey
	sessionSecret := y.tokenInfo.SessionSecret
	if isFamily {
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

func (y *Cloud189PC) EncryptParams(params Params, isFamily bool) string {
	sessionSecret := y.tokenInfo.SessionSecret
	if isFamily {
		sessionSecret = y.tokenInfo.FamilySessionSecret
	}
	if params != nil {
		return AesECBEncrypt(params.Encode(), sessionSecret[:16])
	}
	return ""
}

func (y *Cloud189PC) request(url, method string, callback base.ReqCallback, params Params, resp interface{}, isFamily ...bool) ([]byte, error) {
	req := y.client.R().SetQueryParams(clientSuffix())

	// 设置params
	paramsData := y.EncryptParams(params, isBool(isFamily...))
	if paramsData != "" {
		req.SetQueryParam("params", paramsData)
	}

	// Signature
	req.SetHeaders(y.SignatureHeader(url, method, paramsData, isBool(isFamily...)))

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
		return y.request(url, method, callback, params, resp, isFamily...)
	}

	// if erron.ErrorCode == "InvalidSessionKey" || erron.Code == "InvalidSessionKey" {
	if strings.Contains(res.String(), "InvalidSessionKey") {
		if err = y.refreshSession(); err != nil {
			return nil, err
		}
		return y.request(url, method, callback, params, resp, isFamily...)
	}

	// 处理错误
	if erron.HasError() {
		return nil, &erron
	}
	return res.Body(), nil
}

func (y *Cloud189PC) get(url string, callback base.ReqCallback, resp interface{}, isFamily ...bool) ([]byte, error) {
	return y.request(url, http.MethodGet, callback, nil, resp, isFamily...)
}

func (y *Cloud189PC) post(url string, callback base.ReqCallback, resp interface{}, isFamily ...bool) ([]byte, error) {
	return y.request(url, http.MethodPost, callback, nil, resp, isFamily...)
}

func (y *Cloud189PC) put(ctx context.Context, url string, headers map[string]string, sign bool, file io.Reader, isFamily bool) ([]byte, error) {
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
		for key, value := range y.SignatureHeader(url, http.MethodPut, "", isFamily) {
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
func (y *Cloud189PC) getFiles(ctx context.Context, fileId string, isFamily bool) ([]model.Obj, error) {
	fullUrl := API_URL
	if isFamily {
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
			if isFamily {
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
		}, &resp, isFamily)
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
func (y *Cloud189PC) StreamUpload(ctx context.Context, dstDir model.Obj, file model.FileStreamer, up driver.UpdateProgress, isFamily bool, overwrite bool) (model.Obj, error) {
	var sliceSize = partSize(file.GetSize())
	count := int(math.Ceil(float64(file.GetSize()) / float64(sliceSize)))
	lastPartSize := file.GetSize() % sliceSize
	if file.GetSize() > 0 && lastPartSize == 0 {
		lastPartSize = sliceSize
	}

	params := Params{
		"parentFolderId": dstDir.GetID(),
		"fileName":       url.QueryEscape(file.GetName()),
		"fileSize":       fmt.Sprint(file.GetSize()),
		"sliceSize":      fmt.Sprint(sliceSize),
		"lazyCheck":      "1",
	}

	fullUrl := UPLOAD_URL
	if isFamily {
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
	}, params, &initMultiUpload, isFamily)
	if err != nil {
		return nil, err
	}

	threadG, upCtx := errgroup.NewGroupWithContext(ctx, y.uploadThread,
		retry.Attempts(3),
		retry.Delay(time.Second),
		retry.DelayType(retry.BackOffDelay))

	fileMd5 := md5.New()
	silceMd5 := md5.New()
	silceMd5Hexs := make([]string, 0, count)

	for i := 1; i <= count; i++ {
		if utils.IsCanceled(upCtx) {
			break
		}

		byteData := make([]byte, sliceSize)
		if i == count {
			byteData = byteData[:lastPartSize]
		}

		// 读取块
		silceMd5.Reset()
		if _, err := io.ReadFull(io.TeeReader(file, io.MultiWriter(fileMd5, silceMd5)), byteData); err != io.EOF && err != nil {
			return nil, err
		}

		// 计算块md5并进行hex和base64编码
		md5Bytes := silceMd5.Sum(nil)
		silceMd5Hexs = append(silceMd5Hexs, strings.ToUpper(hex.EncodeToString(md5Bytes)))
		partInfo := fmt.Sprintf("%d-%s", i, base64.StdEncoding.EncodeToString(md5Bytes))

		threadG.Go(func(ctx context.Context) error {
			uploadUrls, err := y.GetMultiUploadUrls(ctx, isFamily, initMultiUpload.Data.UploadFileID, partInfo)
			if err != nil {
				return err
			}

			// step.4 上传切片
			uploadUrl := uploadUrls[0]
			_, err = y.put(ctx, uploadUrl.RequestURL, uploadUrl.Headers, false, bytes.NewReader(byteData), isFamily)
			if err != nil {
				return err
			}
			up(float64(threadG.Success()) * 100 / float64(count))
			return nil
		})
	}
	if err = threadG.Wait(); err != nil {
		return nil, err
	}

	fileMd5Hex := strings.ToUpper(hex.EncodeToString(fileMd5.Sum(nil)))
	sliceMd5Hex := fileMd5Hex
	if file.GetSize() > sliceSize {
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
			"opertype":     IF(overwrite, "3", "1"),
		}, &resp, isFamily)
	if err != nil {
		return nil, err
	}
	return resp.toFile(), nil
}

func (y *Cloud189PC) RapidUpload(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, isFamily bool, overwrite bool) (model.Obj, error) {
	fileMd5 := stream.GetHash().GetHash(utils.MD5)
	if len(fileMd5) < utils.MD5.Width {
		return nil, errors.New("invalid hash")
	}

	uploadInfo, err := y.OldUploadCreate(ctx, dstDir.GetID(), fileMd5, stream.GetName(), fmt.Sprint(stream.GetSize()), isFamily)
	if err != nil {
		return nil, err
	}

	if uploadInfo.FileDataExists != 1 {
		return nil, errors.New("rapid upload fail")
	}

	return y.OldUploadCommit(ctx, uploadInfo.FileCommitUrl, uploadInfo.UploadFileId, isFamily, overwrite)
}

// 快传
func (y *Cloud189PC) FastUpload(ctx context.Context, dstDir model.Obj, file model.FileStreamer, up driver.UpdateProgress, isFamily bool, overwrite bool) (model.Obj, error) {
	tempFile, err := file.CacheFullInTempFile()
	if err != nil {
		return nil, err
	}

	var sliceSize = partSize(file.GetSize())
	count := int(math.Ceil(float64(file.GetSize()) / float64(sliceSize)))
	lastSliceSize := file.GetSize() % sliceSize
	if file.GetSize() > 0 && lastSliceSize == 0 {
		lastSliceSize = sliceSize
	}

	//step.1 优先计算所需信息
	byteSize := sliceSize
	fileMd5 := md5.New()
	silceMd5 := md5.New()
	silceMd5Hexs := make([]string, 0, count)
	partInfos := make([]string, 0, count)
	for i := 1; i <= count; i++ {
		if utils.IsCanceled(ctx) {
			return nil, ctx.Err()
		}

		if i == count {
			byteSize = lastSliceSize
		}

		silceMd5.Reset()
		if _, err := utils.CopyWithBufferN(io.MultiWriter(fileMd5, silceMd5), tempFile, byteSize); err != nil && err != io.EOF {
			return nil, err
		}
		md5Byte := silceMd5.Sum(nil)
		silceMd5Hexs = append(silceMd5Hexs, strings.ToUpper(hex.EncodeToString(md5Byte)))
		partInfos = append(partInfos, fmt.Sprint(i, "-", base64.StdEncoding.EncodeToString(md5Byte)))
	}

	fileMd5Hex := strings.ToUpper(hex.EncodeToString(fileMd5.Sum(nil)))
	sliceMd5Hex := fileMd5Hex
	if file.GetSize() > sliceSize {
		sliceMd5Hex = strings.ToUpper(utils.GetMD5EncodeStr(strings.Join(silceMd5Hexs, "\n")))
	}

	fullUrl := UPLOAD_URL
	if isFamily {
		fullUrl += "/family"
	} else {
		//params.Set("extend", `{"opScene":"1","relativepath":"","rootfolderid":""}`)
		fullUrl += "/person"
	}

	// 尝试恢复进度
	uploadProgress, ok := base.GetUploadProgress[*UploadProgress](y, y.tokenInfo.SessionKey, fileMd5Hex)
	if !ok {
		//step.2 预上传
		params := Params{
			"parentFolderId": dstDir.GetID(),
			"fileName":       url.QueryEscape(file.GetName()),
			"fileSize":       fmt.Sprint(file.GetSize()),
			"fileMd5":        fileMd5Hex,
			"sliceSize":      fmt.Sprint(sliceSize),
			"sliceMd5":       sliceMd5Hex,
		}
		if isFamily {
			params.Set("familyId", y.FamilyID)
		}
		var uploadInfo InitMultiUploadResp
		_, err = y.request(fullUrl+"/initMultiUpload", http.MethodGet, func(req *resty.Request) {
			req.SetContext(ctx)
		}, params, &uploadInfo, isFamily)
		if err != nil {
			return nil, err
		}
		uploadProgress = &UploadProgress{
			UploadInfo:  uploadInfo,
			UploadParts: partInfos,
		}
	}

	uploadInfo := uploadProgress.UploadInfo.Data
	// 网盘中不存在该文件，开始上传
	if uploadInfo.FileDataExists != 1 {
		threadG, upCtx := errgroup.NewGroupWithContext(ctx, y.uploadThread,
			retry.Attempts(3),
			retry.Delay(time.Second),
			retry.DelayType(retry.BackOffDelay))
		for i, uploadPart := range uploadProgress.UploadParts {
			if utils.IsCanceled(upCtx) {
				break
			}

			i, uploadPart := i, uploadPart
			threadG.Go(func(ctx context.Context) error {
				// step.3 获取上传链接
				uploadUrls, err := y.GetMultiUploadUrls(ctx, isFamily, uploadInfo.UploadFileID, uploadPart)
				if err != nil {
					return err
				}
				uploadUrl := uploadUrls[0]

				byteSize, offset := sliceSize, int64(uploadUrl.PartNumber-1)*sliceSize
				if uploadUrl.PartNumber == count {
					byteSize = lastSliceSize
				}

				// step.4 上传切片
				_, err = y.put(ctx, uploadUrl.RequestURL, uploadUrl.Headers, false, io.NewSectionReader(tempFile, offset, byteSize), isFamily)
				if err != nil {
					return err
				}

				up(float64(threadG.Success()) * 100 / float64(len(uploadUrls)))
				uploadProgress.UploadParts[i] = ""
				return nil
			})
		}
		if err = threadG.Wait(); err != nil {
			if errors.Is(err, context.Canceled) {
				uploadProgress.UploadParts = utils.SliceFilter(uploadProgress.UploadParts, func(s string) bool { return s != "" })
				base.SaveUploadProgress(y, uploadProgress, y.tokenInfo.SessionKey, fileMd5Hex)
			}
			return nil, err
		}
	}

	// step.5 提交
	var resp CommitMultiUploadFileResp
	_, err = y.request(fullUrl+"/commitMultiUploadFile", http.MethodGet,
		func(req *resty.Request) {
			req.SetContext(ctx)
		}, Params{
			"uploadFileId": uploadInfo.UploadFileID,
			"isLog":        "0",
			"opertype":     IF(overwrite, "3", "1"),
		}, &resp, isFamily)
	if err != nil {
		return nil, err
	}
	return resp.toFile(), nil
}

// 获取上传切片信息
// 对http body有大小限制，分片信息太多会出错
func (y *Cloud189PC) GetMultiUploadUrls(ctx context.Context, isFamily bool, uploadFileId string, partInfo ...string) ([]UploadUrlInfo, error) {
	fullUrl := UPLOAD_URL
	if isFamily {
		fullUrl += "/family"
	} else {
		fullUrl += "/person"
	}

	var uploadUrlsResp UploadUrlsResp
	_, err := y.request(fullUrl+"/getMultiUploadUrls", http.MethodGet,
		func(req *resty.Request) {
			req.SetContext(ctx)
		}, Params{
			"uploadFileId": uploadFileId,
			"partInfo":     strings.Join(partInfo, ","),
		}, &uploadUrlsResp, isFamily)
	if err != nil {
		return nil, err
	}
	uploadUrls := uploadUrlsResp.Data

	if len(uploadUrls) != len(partInfo) {
		return nil, fmt.Errorf("uploadUrls get error, due to get length %d, real length %d", len(partInfo), len(uploadUrls))
	}

	uploadUrlInfos := make([]UploadUrlInfo, 0, len(uploadUrls))
	for k, uploadUrl := range uploadUrls {
		partNumber, err := strconv.Atoi(strings.TrimPrefix(k, "partNumber_"))
		if err != nil {
			return nil, err
		}
		uploadUrlInfos = append(uploadUrlInfos, UploadUrlInfo{
			PartNumber:     partNumber,
			Headers:        ParseHttpHeader(uploadUrl.RequestHeader),
			UploadUrlsData: uploadUrl,
		})
	}
	sort.Slice(uploadUrlInfos, func(i, j int) bool {
		return uploadUrlInfos[i].PartNumber < uploadUrlInfos[j].PartNumber
	})
	return uploadUrlInfos, nil
}

// 旧版本上传，家庭云不支持覆盖
func (y *Cloud189PC) OldUpload(ctx context.Context, dstDir model.Obj, file model.FileStreamer, up driver.UpdateProgress, isFamily bool, overwrite bool) (model.Obj, error) {
	tempFile, err := file.CacheFullInTempFile()
	if err != nil {
		return nil, err
	}
	fileMd5, err := utils.HashFile(utils.MD5, tempFile)
	if err != nil {
		return nil, err
	}

	// 创建上传会话
	uploadInfo, err := y.OldUploadCreate(ctx, dstDir.GetID(), fileMd5, file.GetName(), fmt.Sprint(file.GetSize()), isFamily)
	if err != nil {
		return nil, err
	}

	// 网盘中不存在该文件，开始上传
	status := GetUploadFileStatusResp{CreateUploadFileResp: *uploadInfo}
	for status.GetSize() < file.GetSize() && status.FileDataExists != 1 {
		if utils.IsCanceled(ctx) {
			return nil, ctx.Err()
		}

		header := map[string]string{
			"ResumePolicy": "1",
			"Expect":       "100-continue",
		}

		if isFamily {
			header["FamilyId"] = fmt.Sprint(y.FamilyID)
			header["UploadFileId"] = fmt.Sprint(status.UploadFileId)
		} else {
			header["Edrive-UploadFileId"] = fmt.Sprint(status.UploadFileId)
		}

		_, err := y.put(ctx, status.FileUploadUrl, header, true, io.NopCloser(tempFile), isFamily)
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
			if isFamily {
				req.SetQueryParam("familyId", fmt.Sprint(y.FamilyID))
			}
		}, &status, isFamily)
		if err != nil {
			return nil, err
		}
		if _, err := tempFile.Seek(status.GetSize(), io.SeekStart); err != nil {
			return nil, err
		}
		up(float64(status.GetSize()) / float64(file.GetSize()) * 100)
	}

	return y.OldUploadCommit(ctx, status.FileCommitUrl, status.UploadFileId, isFamily, overwrite)
}

// 创建上传会话
func (y *Cloud189PC) OldUploadCreate(ctx context.Context, parentID string, fileMd5, fileName, fileSize string, isFamily bool) (*CreateUploadFileResp, error) {
	var uploadInfo CreateUploadFileResp

	fullUrl := API_URL + "/createUploadFile.action"
	if isFamily {
		fullUrl = API_URL + "/family/file/createFamilyFile.action"
	}
	_, err := y.post(fullUrl, func(req *resty.Request) {
		req.SetContext(ctx)
		if isFamily {
			req.SetQueryParams(map[string]string{
				"familyId":     y.FamilyID,
				"parentId":     parentID,
				"fileMd5":      fileMd5,
				"fileName":     fileName,
				"fileSize":     fileSize,
				"resumePolicy": "1",
			})
		} else {
			req.SetFormData(map[string]string{
				"parentFolderId": parentID,
				"fileName":       fileName,
				"size":           fileSize,
				"md5":            fileMd5,
				"opertype":       "3",
				"flag":           "1",
				"resumePolicy":   "1",
				"isLog":          "0",
			})
		}
	}, &uploadInfo, isFamily)

	if err != nil {
		return nil, err
	}
	return &uploadInfo, nil
}

// 提交上传文件
func (y *Cloud189PC) OldUploadCommit(ctx context.Context, fileCommitUrl string, uploadFileID int64, isFamily bool, overwrite bool) (model.Obj, error) {
	var resp OldCommitUploadFileResp
	_, err := y.post(fileCommitUrl, func(req *resty.Request) {
		req.SetContext(ctx)
		if isFamily {
			req.SetHeaders(map[string]string{
				"ResumePolicy": "1",
				"UploadFileId": fmt.Sprint(uploadFileID),
				"FamilyId":     fmt.Sprint(y.FamilyID),
			})
		} else {
			req.SetFormData(map[string]string{
				"opertype":     IF(overwrite, "3", "1"),
				"resumePolicy": "1",
				"uploadFileId": fmt.Sprint(uploadFileID),
				"isLog":        "0",
			})
		}
	}, &resp, isFamily)
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

// 创建家庭云中转文件夹
func (y *Cloud189PC) createFamilyTransferFolder(count int) (*ring.Ring, error) {
	folders := ring.New(count)
	var rootFolder Cloud189Folder
	_, err := y.post(API_URL+"/family/file/createFolder.action", func(req *resty.Request) {
		req.SetQueryParams(map[string]string{
			"folderName": "FamilyTransferFolder",
			"familyId":   y.FamilyID,
		})
	}, &rootFolder, true)
	if err != nil {
		return nil, err
	}

	folderCount := 0

	// 获取已有目录
	files, err := y.getFiles(context.TODO(), rootFolder.GetID(), true)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if folder, ok := file.(*Cloud189Folder); ok {
			folders.Value = folder
			folders = folders.Next()
			folderCount++
		}
	}

	// 创建新的目录
	for folderCount < count {
		var newFolder Cloud189Folder
		_, err := y.post(API_URL+"/family/file/createFolder.action", func(req *resty.Request) {
			req.SetQueryParams(map[string]string{
				"folderName": uuid.NewString(),
				"familyId":   y.FamilyID,
				"parentId":   rootFolder.GetID(),
			})
		}, &newFolder, true)
		if err != nil {
			return nil, err
		}
		folders.Value = &newFolder
		folders = folders.Next()
		folderCount++
	}
	return folders, nil
}

// 清理中转文件夹
func (y *Cloud189PC) cleanFamilyTransfer(ctx context.Context) error {
	var tasks []BatchTaskInfo
	r := y.familyTransferFolder
	for p := r.Next(); p != r; p = p.Next() {
		folder := p.Value.(*Cloud189Folder)

		files, err := y.getFiles(ctx, folder.GetID(), true)
		if err != nil {
			return err
		}
		for _, file := range files {
			tasks = append(tasks, BatchTaskInfo{
				FileId:   file.GetID(),
				FileName: file.GetName(),
				IsFolder: BoolToNumber(file.IsDir()),
			})
		}
	}

	if len(tasks) > 0 {
		// 删除
		resp, err := y.CreateBatchTask("DELETE", y.FamilyID, "", nil, tasks...)
		if err != nil {
			return err
		}
		err = y.WaitBatchTask("DELETE", resp.TaskID, time.Second)
		if err != nil {
			return err
		}
		// 永久删除
		resp, err = y.CreateBatchTask("CLEAR_RECYCLE", y.FamilyID, "", nil, tasks...)
		if err != nil {
			return err
		}
		err = y.WaitBatchTask("CLEAR_RECYCLE", resp.TaskID, time.Second)
		return err
	}
	return nil
}

// 获取家庭云所有用户信息
func (y *Cloud189PC) getFamilyInfoList() ([]FamilyInfoResp, error) {
	var resp FamilyInfoListResp
	_, err := y.get(API_URL+"/family/manage/getFamilyList.action", nil, &resp, true)
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

// 保存家庭云中的文件到个人云
func (y *Cloud189PC) SaveFamilyFileToPersonCloud(ctx context.Context, familyId string, srcObj, dstDir model.Obj, overwrite bool) error {
	// _, err := y.post(API_URL+"/family/file/saveFileToMember.action", func(req *resty.Request) {
	// 	req.SetQueryParams(map[string]string{
	// 		"channelId":    "home",
	// 		"familyId":     familyId,
	// 		"destParentId": destParentId,
	// 		"fileIdList":   familyFileId,
	// 	})
	// }, nil)
	// return err

	task := BatchTaskInfo{
		FileId:   srcObj.GetID(),
		FileName: srcObj.GetName(),
		IsFolder: BoolToNumber(srcObj.IsDir()),
	}
	resp, err := y.CreateBatchTask("COPY", familyId, dstDir.GetID(), map[string]string{
		"groupId":  "null",
		"copyType": "2",
		"shareId":  "null",
	}, task)
	if err != nil {
		return err
	}

	for {
		state, err := y.CheckBatchTask("COPY", resp.TaskID)
		if err != nil {
			return err
		}
		switch state.TaskStatus {
		case 2:
			task.DealWay = IF(overwrite, 3, 2)
			// 冲突时覆盖文件
			if err := y.ManageBatchTask("COPY", resp.TaskID, dstDir.GetID(), task); err != nil {
				return err
			}
		case 4:
			return nil
		}
		time.Sleep(time.Millisecond * 400)
	}
}

func (y *Cloud189PC) CreateBatchTask(aType string, familyID string, targetFolderId string, other map[string]string, taskInfos ...BatchTaskInfo) (*CreateBatchTaskResp, error) {
	var resp CreateBatchTaskResp
	_, err := y.post(API_URL+"/batch/createBatchTask.action", func(req *resty.Request) {
		req.SetFormData(map[string]string{
			"type":      aType,
			"taskInfos": MustString(utils.Json.MarshalToString(taskInfos)),
		})
		if targetFolderId != "" {
			req.SetFormData(map[string]string{"targetFolderId": targetFolderId})
		}
		if familyID != "" {
			req.SetFormData(map[string]string{"familyId": familyID})
		}
		req.SetFormData(other)
	}, &resp, familyID != "")
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// 检测任务状态
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

// 获取冲突的任务信息
func (y *Cloud189PC) GetConflictTaskInfo(aType string, taskID string) (*BatchTaskConflictTaskInfoResp, error) {
	var resp BatchTaskConflictTaskInfoResp
	_, err := y.post(API_URL+"/batch/getConflictTaskInfo.action", func(req *resty.Request) {
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

// 处理冲突
func (y *Cloud189PC) ManageBatchTask(aType string, taskID string, targetFolderId string, taskInfos ...BatchTaskInfo) error {
	_, err := y.post(API_URL+"/batch/manageBatchTask.action", func(req *resty.Request) {
		req.SetFormData(map[string]string{
			"targetFolderId": targetFolderId,
			"type":           aType,
			"taskId":         taskID,
			"taskInfos":      MustString(utils.Json.MarshalToString(taskInfos)),
		})
	}, nil)
	return err
}

var ErrIsConflict = errors.New("there is a conflict with the target object")

// 等待任务完成
func (y *Cloud189PC) WaitBatchTask(aType string, taskID string, t time.Duration) error {
	for {
		state, err := y.CheckBatchTask(aType, taskID)
		if err != nil {
			return err
		}
		switch state.TaskStatus {
		case 2:
			return ErrIsConflict
		case 4:
			return nil
		}
		time.Sleep(t)
	}
}
