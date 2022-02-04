package _89

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var client189Map map[string]*resty.Client
var infoMap = make(map[string]Rsa)

func (driver Cloud189) getClient(account *model.Account) (*resty.Client, error) {
	client, ok := client189Map[account.Name]
	if ok {
		return client, nil
	}
	err := driver.Login(account)
	if err != nil {
		return nil, err
	}
	client, ok = client189Map[account.Name]
	if !ok {
		return nil, fmt.Errorf("can't find [%s] client", account.Name)
	}
	return client, nil
}

func (driver Cloud189) FormatFile(file *Cloud189File) *model.File {
	f := &model.File{
		Id:        strconv.FormatInt(file.Id, 10),
		Name:      file.Name,
		Size:      file.Size,
		Driver:    driver.Config().Name,
		UpdatedAt: nil,
		Thumbnail: file.Icon.SmallUrl,
		Url:       file.Url,
	}
	loc, _ := time.LoadLocation("Local")
	lastOpTime, err := time.ParseInLocation("2006-01-02 15:04:05", file.LastOpTime, loc)
	if err == nil {
		f.UpdatedAt = &lastOpTime
	}
	if file.Size == -1 {
		f.Type = conf.FOLDER
		f.Size = 0
	} else {
		f.Type = utils.GetFileType(filepath.Ext(file.Name))
	}
	return f
}

//func (c Cloud189) GetFile(path string, account *model.Account) (*Cloud189File, error) {
//	dir, name := filepath.Split(path)
//	dir = utils.ParsePath(dir)
//	_, _, err := c.ParentPath(dir, account)
//	if err != nil {
//		return nil, err
//	}
//	parentFiles_, _ := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, dir))
//	parentFiles, _ := parentFiles_.([]Cloud189File)
//	for _, file := range parentFiles {
//		if file.Name == name {
//			if file.Size != -1 {
//				return &file, err
//			} else {
//				return nil, ErrNotFile
//			}
//		}
//	}
//	return nil, ErrPathNotFound
//}

type Cloud189Down struct {
	ResCode         int    `json:"res_code"`
	ResMessage      string `json:"res_message"`
	FileDownloadUrl string `json:"fileDownloadUrl"`
}

type LoginResp struct {
	Msg    string `json:"msg"`
	Result int    `json:"result"`
	ToUrl  string `json:"toUrl"`
}

// Login refer to PanIndex
func (driver Cloud189) Login(account *model.Account) error {
	client, ok := client189Map[account.Name]
	if !ok {
		//cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		client = resty.New()
		//client.SetCookieJar(cookieJar)
		client.SetRetryCount(3)
		client.SetHeader("Referer", "https://cloud.189.cn/")
	}
	url := "https://cloud.189.cn/api/portal/loginUrl.action?redirectURL=https%3A%2F%2Fcloud.189.cn%2Fmain.action"
	b := ""
	lt := ""
	ltText := regexp.MustCompile(`lt = "(.+?)"`)
	for i := 0; i < 3; i++ {
		res, err := client.R().Get(url)
		if err != nil {
			return err
		}
		// 已经登陆
		if res.RawResponse.Request.URL.String() == "https://cloud.189.cn/web/main" {
			return nil
		}
		b = res.String()
		ltTextArr := ltText.FindStringSubmatch(b)
		if len(ltTextArr) > 0 {
			lt = ltTextArr[1]
			break
		} else {
			<-time.After(time.Second)
		}
	}
	if lt == "" {
		return fmt.Errorf("get empty login page")
	}
	captchaToken := regexp.MustCompile(`captchaToken' value='(.+?)'`).FindStringSubmatch(b)[1]
	returnUrl := regexp.MustCompile(`returnUrl = '(.+?)'`).FindStringSubmatch(b)[1]
	paramId := regexp.MustCompile(`paramId = "(.+?)"`).FindStringSubmatch(b)[1]
	//reqId := regexp.MustCompile(`reqId = "(.+?)"`).FindStringSubmatch(b)[1]
	jRsakey := regexp.MustCompile(`j_rsaKey" value="(\S+)"`).FindStringSubmatch(b)[1]
	vCodeID := regexp.MustCompile(`picCaptcha\.do\?token\=([A-Za-z0-9\&\=]+)`).FindStringSubmatch(b)[1]
	vCodeRS := ""
	if vCodeID != "" {
		// need ValidateCode
		log.Debugf("try to identify verification codes")
		timeStamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
		u := "https://open.e.189.cn/api/logbox/oauth2/picCaptcha.do?token=" + vCodeID + timeStamp
		imgRes, err := client.R().SetHeaders(map[string]string{
			"User-Agent":     "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:74.0) Gecko/20100101 Firefox/76.0",
			"Referer":        "https://open.e.189.cn/api/logbox/oauth2/unifyAccountLogin.do",
			"Sec-Fetch-Dest": "image",
			"Sec-Fetch-Mode": "no-cors",
			"Sec-Fetch-Site": "same-origin",
		}).Get(u)
		if err != nil {
			return err
		}
		vRes, err := client.R().SetMultipartField(
			"image", "validateCode.png", "image/png", bytes.NewReader(imgRes.Body())).
			Post(conf.GetStr("ocr api"))
		if err != nil {
			return err
		}
		if jsoniter.Get(vRes.Body(), "status").ToInt() != 200 {
			return errors.New("ocr error:" + jsoniter.Get(vRes.Body(), "msg").ToString())
		}
		vCodeRS = jsoniter.Get(vRes.Body(), "result").ToString()
		log.Debugln("code: ", vCodeRS)
	}
	userRsa := RsaEncode([]byte(account.Username), jRsakey, true)
	passwordRsa := RsaEncode([]byte(account.Password), jRsakey, true)
	url = "https://open.e.189.cn/api/logbox/oauth2/loginSubmit.do"
	var loginResp LoginResp
	res, err := client.R().
		SetHeaders(map[string]string{
			"lt":         lt,
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36",
			"Referer":    "https://open.e.189.cn/",
			"accept":     "application/json;charset=UTF-8",
		}).SetFormData(map[string]string{
		"appKey":       "cloud",
		"accountType":  "01",
		"userName":     "{RSA}" + userRsa,
		"password":     "{RSA}" + passwordRsa,
		"validateCode": vCodeRS,
		"captchaToken": captchaToken,
		"returnUrl":    returnUrl,
		"mailSuffix":   "@pan.cn",
		"paramId":      paramId,
		"clientType":   "10010",
		"dynamicCheck": "FALSE",
		"cb_SaveName":  "1",
		"isOauth2":     "false",
	}).Post(url)
	if err != nil {
		return err
	}
	err = utils.Json.Unmarshal(res.Body(), &loginResp)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if loginResp.Result != 0 {
		return fmt.Errorf(loginResp.Msg)
	}
	_, err = client.R().Get(loginResp.ToUrl)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	client189Map[account.Name] = client
	return nil
}

func (driver Cloud189) isFamily(account *model.Account) bool {
	return account.InternalType == "Family"
}

func (driver Cloud189) GetFiles(fileId string, account *model.Account) ([]Cloud189File, error) {
	res := make([]Cloud189File, 0)
	pageNum := 1

	for {
		var resp Cloud189Files
		body, err := driver.Request("https://cloud.189.cn/api/open/file/listFiles.action", base.Get, map[string]string{
			//"noCache":    random(),
			"pageSize":   "60",
			"pageNum":    strconv.Itoa(pageNum),
			"mediaType":  "0",
			"folderId":   fileId,
			"iconOption": "5",
			"orderBy":    "lastOpTime", //account.OrderBy
			"descending": "true",       //account.OrderDirection
		}, nil, nil, account)
		if err != nil {
			return nil, err
		}
		err = utils.Json.Unmarshal(body, &resp)
		if err != nil {
			return nil, err
		}
		if resp.ResCode != 0 {
			return nil, fmt.Errorf(resp.ResMessage)
		}
		if resp.FileListAO.Count == 0 {
			break
		}
		for _, folder := range resp.FileListAO.FolderList {
			res = append(res, Cloud189File{
				Id:         folder.Id,
				LastOpTime: folder.LastOpTime,
				Name:       folder.Name,
				Size:       -1,
			})
		}
		res = append(res, resp.FileListAO.FileList...)
		pageNum++
	}
	return res, nil
}

func (driver Cloud189) Request(url string, method int, query, form map[string]string, headers map[string]string, account *model.Account) ([]byte, error) {
	client, err := driver.getClient(account)
	if err != nil {
		return nil, err
	}
	//var resp base.Json
	if driver.isFamily(account) {
		url = strings.Replace(url, "/api/open", "/api/open/family", 1)
		if query != nil {
			query["familyId"] = account.SiteId
		}
		if form != nil {
			form["familyId"] = account.SiteId
		}
	}
	var e Cloud189Error
	req := client.R().SetError(&e).
		SetHeader("Accept", "application/json;charset=UTF-8").
		SetQueryParams(map[string]string{
			"noCache": random(),
		})
	if query != nil {
		req = req.SetQueryParams(query)
	}
	if form != nil {
		req = req.SetFormData(form)
	}
	if headers != nil {
		req = req.SetHeaders(headers)
	}
	var res *resty.Response
	switch method {
	case base.Get:
		res, err = req.Get(url)
	case base.Post:
		res, err = req.Post(url)
	default:
		return nil, base.ErrNotSupport
	}
	if err != nil {
		return nil, err
	}
	//log.Debug(res.String())
	if e.ErrorCode != "" {
		if e.ErrorCode == "InvalidSessionKey" {
			err = driver.Login(account)
			if err != nil {
				return nil, err
			}
			return driver.Request(url, method, query, form, nil, account)
		}
	}
	if jsoniter.Get(res.Body(), "res_code").ToInt() != 0 {
		err = errors.New(jsoniter.Get(res.Body(), "res_message").ToString())
	}
	return res.Body(), err
}

func (driver Cloud189) GetSessionKey(account *model.Account) (string, error) {
	//info, ok := infoMap[account.Name]
	//if !ok {
	//	info = Info{}
	//	infoMap[account.Name] = info
	//} else {
	//	log.Debugf("hit")
	//}
	//if info.SessionKey != "" {
	//	return info.SessionKey, nil
	//}
	resp, err := driver.Request("https://cloud.189.cn/v2/getUserBriefInfo.action", base.Get, nil, nil, nil, account)
	if err != nil {
		return "", err
	}
	sessionKey := jsoniter.Get(resp, "sessionKey").ToString()
	//info.SessionKey = sessionKey
	return sessionKey, nil
}

func (driver Cloud189) GetResKey(account *model.Account) (string, string, error) {
	rsa, ok := infoMap[account.Name]
	if !ok {
		rsa = Rsa{}
		infoMap[account.Name] = rsa
	}
	now := time.Now().UnixMilli()
	if rsa.Expire > now {
		return rsa.PubKey, rsa.PkId, nil
	}
	resp, err := driver.Request("https://cloud.189.cn/api/security/generateRsaKey.action", base.Get, nil, nil, nil, account)
	if err != nil {
		return "", "", err
	}
	pubKey, pkId := jsoniter.Get(resp, "pubKey").ToString(), jsoniter.Get(resp, "pkId").ToString()
	rsa.PubKey, rsa.PkId = pubKey, pkId
	rsa.Expire = jsoniter.Get(resp, "expire").ToInt64()
	return pubKey, pkId, nil
}

//func (driver Cloud189) UploadRequest1(uri string, form map[string]string, account *model.Account, resp interface{}) ([]byte, error) {
//	//sessionKey, err := driver.GetSessionKey(account)
//	//if err != nil {
//	//	return nil, err
//	//}
//	sessionKey := account.DriveId
//	pubKey, pkId, err := driver.GetResKey(account)
//	log.Debugln(sessionKey, pubKey, pkId)
//	if err != nil {
//		return nil, err
//	}
//	xRId := Random("xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx")
//	pkey := Random("xxxxxxxxxxxx4xxxyxxxxxxxxxxxxxxx")[0 : 16+int(16*mathRand.Float32())]
//	params := hex.EncodeToString(AesEncrypt([]byte(qs(form)), []byte(pkey[0:16])))
//	date := strconv.FormatInt(time.Now().Unix(), 10)
//	a := make(url.Values)
//	a.Set("SessionKey", sessionKey)
//	a.Set("Operate", http.MethodGet)
//	a.Set("RequestURI", uri)
//	a.Set("Date", date)
//	a.Set("params", params)
//	signature := hex.EncodeToString(SHA1(EncodeParam(a), pkey))
//	encryptionText := RsaEncode([]byte(pkey), pubKey, false)
//	headers := map[string]string{
//		"signature":      signature,
//		"sessionKey":     sessionKey,
//		"encryptionText": encryptionText,
//		"pkId":           pkId,
//		"x-request-id":   xRId,
//		"x-request-date": date,
//	}
//	req := base.RestyClient.R().SetHeaders(headers).SetQueryParam("params", params)
//	if resp != nil {
//		req.SetResult(resp)
//	}
//	res, err := req.Get("https://upload.cloud.189.cn" + uri)
//	if err != nil {
//		return nil, err
//	}
//	//log.Debug(res.String())
//	data := res.Body()
//	if jsoniter.Get(data, "code").ToString() != "SUCCESS" {
//		return nil, errors.New(uri + "---" + jsoniter.Get(data, "msg").ToString())
//	}
//	return data, nil
//}
//
//func (driver Cloud189) UploadRequest2(uri string, form map[string]string, account *model.Account, resp interface{}) ([]byte, error) {
//	c := strconv.FormatInt(time.Now().UnixMilli(), 10)
//	r := Random("xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx")
//	l := Random("xxxxxxxxxxxx4xxxyxxxxxxxxxxxxxxx")
//	l = l[0 : 16+int(16*mathRand.Float32())]
//
//	e := qs(form)
//	data := AesEncrypt([]byte(e), []byte(l[0:16]))
//	h := hex.EncodeToString(data)
//
//	sessionKey := account.DriveId
//	a := make(url.Values)
//	a.Set("SessionKey", sessionKey)
//	a.Set("Operate", http.MethodGet)
//	a.Set("RequestURI", uri)
//	a.Set("Date", c)
//	a.Set("params", h)
//	g := SHA1(EncodeParam(a), l)
//
//	pubKey, pkId, err := driver.GetResKey(account)
//	if err != nil {
//		return nil, err
//	}
//	b := RsaEncode([]byte(l), pubKey, false)
//	client, err := driver.getClient(account)
//	if err != nil {
//		return nil, err
//	}
//	req := client.R()
//	req.Header.Set("accept", "application/json;charset=UTF-8")
//	req.Header.Set("SessionKey", sessionKey)
//	req.Header.Set("Signature", hex.EncodeToString(g))
//	req.Header.Set("X-Request-Date", c)
//	req.Header.Set("X-Request-ID", r)
//	req.Header.Set("EncryptionText", b)
//	req.Header.Set("PkId", pkId)
//	if resp != nil {
//		req.SetResult(resp)
//	}
//	res, err := req.Get("https://upload.cloud.189.cn" + uri + "?params=" + h)
//	if err != nil {
//		return nil, err
//	}
//	//log.Debug(res.String())
//	data = res.Body()
//	if jsoniter.Get(data, "code").ToString() != "SUCCESS" {
//		return nil, errors.New(uri + "---" + jsoniter.Get(data, "msg").ToString())
//	}
//	return data, nil
//}

func (driver Cloud189) UploadRequest(uri string, form map[string]string, account *model.Account, resp interface{}) ([]byte, error) {
	c := strconv.FormatInt(time.Now().UnixMilli(), 10)
	r := Random("xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx")
	l := Random("xxxxxxxxxxxx4xxxyxxxxxxxxxxxxxxx")
	l = l[0 : 16+int(16*utils.Rand.Float32())]

	e := qs(form)
	data := AesEncrypt([]byte(e), []byte(l[0:16]))
	h := hex.EncodeToString(data)

	sessionKey := account.DriveId
	signature := hmacSha1(fmt.Sprintf("SessionKey=%s&Operate=GET&RequestURI=%s&Date=%s&params=%s", sessionKey, uri, c, h), l)

	pubKey, pkId, err := driver.GetResKey(account)
	if err != nil {
		return nil, err
	}
	b := RsaEncode([]byte(l), pubKey, false)
	client, err := driver.getClient(account)
	if err != nil {
		return nil, err
	}
	req := client.R()
	req.Header.Set("accept", "application/json;charset=UTF-8")
	req.Header.Set("SessionKey", sessionKey)
	req.Header.Set("Signature", signature)
	req.Header.Set("X-Request-Date", c)
	req.Header.Set("X-Request-ID", r)
	req.Header.Set("EncryptionText", b)
	req.Header.Set("PkId", pkId)
	if resp != nil {
		req.SetResult(resp)
	}
	res, err := req.Get("https://upload.cloud.189.cn" + uri + "?params=" + h)
	if err != nil {
		return nil, err
	}
	//log.Debug(res.String())
	data = res.Body()
	if jsoniter.Get(data, "code").ToString() != "SUCCESS" {
		return nil, errors.New(uri + "---" + jsoniter.Get(data, "msg").ToString())
	}
	return data, nil
}

// NewUpload Error: signature check false
func (driver Cloud189) NewUpload(file *model.FileStream, account *model.Account) error {
	sessionKey, err := driver.GetSessionKey(account)
	if err != nil {
		account.Status = err.Error()
	} else {
		account.Status = "work"
		account.DriveId = sessionKey
	}
	_ = model.SaveAccount(account)
	const DEFAULT uint64 = 10485760
	var count = int64(math.Ceil(float64(file.GetSize()) / float64(DEFAULT)))
	var finish uint64 = 0
	parentFile, err := driver.File(file.ParentPath, account)
	if err != nil {
		return err
	}
	if !parentFile.IsDir() {
		return base.ErrNotFolder
	}
	res, err := driver.UploadRequest("/person/initMultiUpload", map[string]string{
		"parentFolderId": parentFile.Id,
		"fileName":       file.Name,
		"fileSize":       strconv.FormatInt(int64(file.Size), 10),
		"sliceSize":      strconv.FormatInt(int64(DEFAULT), 10),
		"lazyCheck":      "1",
	}, account, nil)
	if err != nil {
		return err
	}
	uploadFileId := jsoniter.Get(res, "data", "uploadFileId").ToString()
	//_, err = driver.UploadRequest("/person/getUploadedPartsInfo", map[string]string{
	//	"uploadFileId": uploadFileId,
	//}, account, nil)
	var i int64
	var byteSize uint64
	md5s := make([]string, 0)
	md5Sum := md5.New()
	for i = 1; i <= count; i++ {
		byteSize = file.GetSize() - finish
		if DEFAULT < byteSize {
			byteSize = DEFAULT
		}
		//log.Debugf("%d,%d", byteSize, finish)
		byteData := make([]byte, byteSize)
		n, err := io.ReadFull(file, byteData)
		//log.Debug(err, n)
		if err != nil {
			return err
		}
		finish += uint64(n)
		md5Bytes := getMd5(byteData)
		md5Hex := hex.EncodeToString(md5Bytes)
		md5Base64 := base64.StdEncoding.EncodeToString(md5Bytes)
		md5s = append(md5s, strings.ToUpper(md5Hex))
		md5Sum.Write(byteData)
		//log.Debugf("md5Bytes: %+v,md5Str:%s,md5Base64:%s", md5Bytes, md5Hex, md5Base64)
		var resp UploadUrlsResp
		res, err = driver.UploadRequest("/person/getMultiUploadUrls", map[string]string{
			"partInfo":     fmt.Sprintf("%s-%s", strconv.FormatInt(i, 10), md5Base64),
			"uploadFileId": uploadFileId,
		}, account, &resp)
		if err != nil {
			return err
		}
		uploadData := resp.UploadUrls["partNumber_"+strconv.FormatInt(i, 10)]
		log.Debugf("uploadData: %+v", uploadData)
		requestURL := uploadData.RequestURL
		uploadHeaders := strings.Split(decodeURIComponent(uploadData.RequestHeader), "&")
		req, _ := http.NewRequest(http.MethodPut, requestURL, bytes.NewReader(byteData))
		for _, v := range uploadHeaders {
			i := strings.Index(v, "=")
			req.Header.Set(v[0:i], v[i+1:])
		}

		r, err := base.HttpClient.Do(req)
		log.Debugf("%+v %+v", r, r.Request.Header)
		if err != nil {
			return err
		}
	}
	fileMd5 := hex.EncodeToString(md5Sum.Sum(nil))
	sliceMd5 := fileMd5
	if file.GetSize() > DEFAULT {
		sliceMd5 = utils.GetMD5Encode(strings.Join(md5s, "\n"))
	}
	res, err = driver.UploadRequest("/person/commitMultiUploadFile", map[string]string{
		"uploadFileId": uploadFileId,
		"fileMd5":      fileMd5,
		"sliceMd5":     sliceMd5,
		"lazyCheck":    "1",
	}, account, nil)
	account.DriveId, _ = driver.GetSessionKey(account)
	return err
}

func (driver Cloud189) OldUpload(file *model.FileStream, account *model.Account) error {
	//return base.ErrNotImplement
	client, err := driver.getClient(account)
	if err != nil {
		return err
	}
	parentFile, err := driver.File(file.ParentPath, account)
	if err != nil {
		return err
	}
	// api refer to PanIndex
	res, err := client.R().SetMultipartFormData(map[string]string{
		"parentId":   parentFile.Id,
		"sessionKey": account.DriveId,
		"opertype":   "1",
		"fname":      file.GetFileName(),
	}).SetMultipartField("Filedata", file.GetFileName(), file.GetMIMEType(), file).Post("https://hb02.upload.cloud.189.cn/v1/DCIWebUploadAction")
	if err != nil {
		return err
	}
	if jsoniter.Get(res.Body(), "MD5").ToString() != "" {
		return nil
	}
	log.Debugf(res.String())
	return errors.New(res.String())
}
