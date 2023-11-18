package vtencent

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

func (d *Vtencent) request(url, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"cookie":       d.Cookie,
		"content-type": "application/json",
		"origin":       d.conf.origin,
		"referer":      d.conf.referer,
	})
	if callback != nil {
		callback(req)
	} else {
		req.SetBody("{}")
	}
	if resp != nil {
		req.SetResult(resp)
	}
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}
	code := utils.Json.Get(res.Body(), "Code").ToString()
	if code != "Success" {
		switch code {
		case "AuthFailure.SessionInvalid":
			if err != nil {
				return nil, errors.New(code)
			}
		default:
			return nil, errors.New(code)
		}
		return d.request(url, method, callback, resp)
	}
	return res.Body(), nil
}

func (d *Vtencent) ugcRequest(url, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"cookie":       d.Cookie,
		"content-type": "application/json",
		"origin":       d.conf.origin,
		"referer":      d.conf.referer,
	})
	if callback != nil {
		callback(req)
	} else {
		req.SetBody("{}")
	}
	if resp != nil {
		req.SetResult(resp)
	}
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}
	code := utils.Json.Get(res.Body(), "Code").ToInt()
	if code != 0 {
		message := utils.Json.Get(res.Body(), "message").ToString()
		if len(message) == 0 {
			message = utils.Json.Get(res.Body(), "msg").ToString()
		}
		return nil, errors.New(message)
	}
	return res.Body(), nil
}

func (d *Vtencent) LoadUser() (string, error) {
	api := "https://api.vs.tencent.com/SaaS/Account/DescribeAccount"
	res, err := d.request(api, http.MethodPost, func(req *resty.Request) {}, nil)
	if err != nil {
		return "", err
	}
	return utils.Json.Get(res, "Data", "TfUid").ToString(), nil
}

func (d *Vtencent) GetFiles(dirId string) ([]File, error) {
	api := "https://api.vs.tencent.com/PaaS/Material/SearchResource"
	form := fmt.Sprintf(`{
		"Text":"",
		"Text":"",
		"Offset":0,
		"Limit":20000,
		"Sort":{"Field":"%s","Order":"%s"},
		"CreateTimeRanges":[],
		"MaterialTypes":[],
		"ReviewStatuses":[],
		"Tags":[],
		"SearchScopes":[{"Owner":{"Type":"PERSON","Id":"%s"},"ClassId":%s,"SearchOneDepth":true}]
	}`, d.Addition.OrderBy, d.Addition.OrderDirection, d.TfUid, dirId)
	var dat map[string]interface{}
	if err := json.Unmarshal([]byte(form), &dat); err != nil {
		return []File{}, err
	}
	var resps RspFiles
	rsp, err := d.request(api, http.MethodPost, func(req *resty.Request) {
		req.SetBody(dat)
	}, &resps)
	if err != nil {
		return []File{}, err
	}
	if err := json.Unmarshal(rsp, &resps); err != nil {
		return []File{}, err
	}
	return resps.Data.ResourceInfoSet, nil
}

func (d *Vtencent) CreateUploadMaterial(classId int, fileName string, UploadSummaryKey string) (RspCreatrMaterial, error) {
	api := "https://api.vs.tencent.com/PaaS/Material/CreateUploadMaterial"
	form := base.Json{"Owner": base.Json{"Type": "PERSON", "Id": d.TfUid},
		"MaterialType": "IMAGE", "Name": fileName, "ClassId": classId,
		"UploadSummaryKey": UploadSummaryKey}
	rsp, err := d.request(api, http.MethodPost, func(req *resty.Request) {
		req.SetBody(form)
	}, nil)
	if err != nil {
		return RspCreatrMaterial{}, err
	}
	var resps RspCreatrMaterial
	if err := json.Unmarshal(rsp, &resps); err != nil {
		return RspCreatrMaterial{}, err
	}
	return resps, nil
}

func (d *Vtencent) ApplyUploadUGC(signature string, stream model.FileStreamer) (RspApplyUploadUGC, error) {
	api := "https://vod2.qcloud.com/v3/index.php?Action=ApplyUploadUGC"
	form := base.Json{
		"signature": signature,
		"videoName": stream.GetName(),
		"videoType": strings.ReplaceAll(path.Ext(stream.GetName()), ".", ""),
		"videoSize": stream.GetSize(),
	}
	rsp, err := d.ugcRequest(api, http.MethodPost, func(req *resty.Request) {
		req.SetBody(form)
	}, nil)
	if err != nil {
		return RspApplyUploadUGC{}, err
	}
	var resps RspApplyUploadUGC
	if err := json.Unmarshal(rsp, &resps); err != nil {
		return RspApplyUploadUGC{}, err
	}
	return resps, nil
}

func (d *Vtencent) CommitUploadUGC(signature string, vodSessionKey string) (RspCommitUploadUGC, error) {
	api := "https://vod2.qcloud.com/v3/index.php?Action=CommitUploadUGC"
	form := base.Json{
		"signature":     signature,
		"vodSessionKey": vodSessionKey,
	}
	rsp, err := d.ugcRequest(api, http.MethodPost, func(req *resty.Request) {
		req.SetBody(form)
	}, nil)
	if err != nil {
		return RspCommitUploadUGC{}, err
	}
	var resps RspCommitUploadUGC
	if err := json.Unmarshal(rsp, &resps); err != nil {
		return RspCommitUploadUGC{}, err
	}
	if len(resps.Data.Video.URL) == 0 {
		return RspCommitUploadUGC{}, errors.New(string(rsp))
	}
	return resps, nil
}

func (d *Vtencent) GetUploadId(rspUGC RspApplyUploadUGC) (string, error) {
	StoragePath := strings.ReplaceAll(rspUGC.Data.Video.StoragePath, "/"+rspUGC.Data.StorageBucket, rspUGC.Data.StorageBucket)
	StoragePath = url.QueryEscape(StoragePath)
	keyTime := fmt.Sprintf("%d;%d", rspUGC.Data.Timestamp, rspUGC.Data.TempCertificate.ExpiredTime)
	signPath := fmt.Sprintf("get\n/\nprefix=%s&uploads=\n\n", StoragePath)
	singleKey := QSignatureKey(keyTime, signPath, rspUGC.Data.TempCertificate.SecretKey)
	authorization := fmt.Sprintf("q-sign-algorithm=sha1&q-ak=%s&q-sign-time=%s&q-key-time=%s&q-header-list=&q-url-param-list=prefix;uploads&q-signature=%s", rspUGC.Data.TempCertificate.SecretID, keyTime, keyTime, singleKey)
	token := rspUGC.Data.TempCertificate.Token
	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"Authorization":        authorization,
		"X-Cos-Security-Token": token,
		"X-Cos-Storage-Class":  "Standard",
	})
	api := fmt.Sprintf("https://%s-%d.cos.%s.myqcloud.com/?uploads&prefix=%s", rspUGC.Data.StorageBucket, rspUGC.Data.StorageAppID, rspUGC.Data.StorageRegionV5, StoragePath)
	res, err := req.Execute("GET", api)
	if err != nil {
		return "", err
	}
	if !strings.Contains(string(res.Body()), "1000") {
		if err != nil {
			return "", errors.New(string(res.Body()))
		}
	}
	signPath = fmt.Sprintf("post\n%s\nuploads=\nx-cos-storage-class=Standard\n", rspUGC.Data.Video.StoragePath)
	singleKeyTwo := QSignatureKey(keyTime, signPath, rspUGC.Data.TempCertificate.SecretKey)
	authorizationTwo := fmt.Sprintf("q-sign-algorithm=sha1&q-ak=%s&q-sign-time=%s&q-key-time=%s&q-header-list=x-cos-storage-class&q-url-param-list=uploads&q-signature=%s", rspUGC.Data.TempCertificate.SecretID, keyTime, keyTime, singleKeyTwo)
	reqt := base.RestyClient.R()
	reqt.SetHeaders(map[string]string{
		"Authorization":        authorizationTwo,
		"X-Cos-Security-Token": token,
		"X-Cos-Storage-Class":  "Standard",
	})
	upIdUrl := fmt.Sprintf("https://%s-%d.cos.%s.myqcloud.com%s", rspUGC.Data.StorageBucket, rspUGC.Data.StorageAppID, rspUGC.Data.StorageRegionV5, rspUGC.Data.Video.StoragePath)
	rest, err := reqt.Execute("POST", upIdUrl+"?uploads")
	if err != nil {
		return "", err
	}
	if !strings.Contains(string(rest.Body()), "UploadId") {
		if err != nil {
			return "", errors.New(string(rest.Body()))
		}
	}
	re := regexp.MustCompile(`<UploadId>(.*?)</UploadId>`)
	matches := re.FindAllStringSubmatch(string(rest.Body()), -1)
	if len(matches) <= 0 {
		return "", errors.New("UploadId not found")
	}
	uploadId := matches[0][1]
	return uploadId, nil
}

func (d *Vtencent) PutFile(uploadId string, rspUGC RspApplyUploadUGC, stream model.FileStreamer) (string, error) {
	keyTime := fmt.Sprintf("%d;%d", rspUGC.Data.Timestamp, rspUGC.Data.TempCertificate.ExpiredTime)
	baseUrl := fmt.Sprintf("https://%s-%d.cos.%s.myqcloud.com%s", rspUGC.Data.StorageBucket, rspUGC.Data.StorageAppID, rspUGC.Data.StorageRegionV5, rspUGC.Data.Video.StoragePath)
	putUrl := baseUrl + "?partNumber=1&uploadId=" + uploadId
	signPath := fmt.Sprintf("put\n%s\npartnumber=1&uploadid=%s\ncontent-length=%d\n", rspUGC.Data.Video.StoragePath, uploadId, int(stream.GetSize()))
	token := rspUGC.Data.TempCertificate.Token
	form, err := io.ReadAll(stream)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("PUT", putUrl, bytes.NewBuffer(form))
	if err != nil {
		return "", err
	}
	host := fmt.Sprintf("%s-%d.cos.%s.myqcloud.com", rspUGC.Data.StorageBucket, rspUGC.Data.StorageAppID, rspUGC.Data.StorageRegionV5)
	singleKeyUpload := QSignatureKey(keyTime, signPath, rspUGC.Data.TempCertificate.SecretKey)
	authorization := fmt.Sprintf("q-sign-algorithm=sha1&q-ak=%s&q-sign-time=%s&q-key-time=%s&q-header-list=content-length&q-url-param-list=partnumber;uploadid&q-signature=%s", rspUGC.Data.TempCertificate.SecretID, keyTime, keyTime, singleKeyUpload)
	req.Header.Set("Content-Length", strconv.Itoa(int(stream.GetSize())))
	req.Header.Set("Authorization", authorization)
	req.Header.Set("X-Cos-Security-Token", token)
	req.Header.Set("Host", host)
	rsp, err := base.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}
	if strings.Contains(string(body), "xml") {
		return "", errors.New(string(body))
	}
	Etag := rsp.Header.Get("Etag")
	return Etag, nil
}

func (d *Vtencent) CompleteMultipartUpload(uploadId string, Etag string, rspUGC RspApplyUploadUGC, stream model.FileStreamer) error {
	keyTime := fmt.Sprintf("%d;%d", rspUGC.Data.Timestamp, rspUGC.Data.TempCertificate.ExpiredTime)
	token := rspUGC.Data.TempCertificate.Token
	form, err := io.ReadAll(stream)
	if err != nil {
		return err
	}
	baseUrl := fmt.Sprintf("https://%s-%d.cos.%s.myqcloud.com%s", rspUGC.Data.StorageBucket, rspUGC.Data.StorageAppID, rspUGC.Data.StorageRegionV5, rspUGC.Data.Video.StoragePath)
	putUrl := baseUrl + "?uploadId=" + uploadId
	hash := md5.Sum(form)
	base64Hash := base64.StdEncoding.EncodeToString(hash[:])
	fileMd5 := url.QueryEscape(base64Hash)
	signPath := fmt.Sprintf("post\n%s\nuploadid=%s\ncontent-md5=%s\n", rspUGC.Data.Video.StoragePath, uploadId, fileMd5)
	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<CompleteMultipartUpload>
  <Part>
    <PartNumber>1</PartNumber>
    <ETag>%s</ETag>
  </Part>
</CompleteMultipartUpload>`, Etag)
	req, err := http.NewRequest("POST", putUrl, bytes.NewBuffer([]byte(xmlData)))
	if err != nil {
		return err
	}
	host := fmt.Sprintf("%s-%d.cos.%s.myqcloud.com", rspUGC.Data.StorageBucket, rspUGC.Data.StorageAppID, rspUGC.Data.StorageRegionV5)
	singleKeyUpload := QSignatureKey(keyTime, signPath, rspUGC.Data.TempCertificate.SecretKey)
	authorization := fmt.Sprintf("q-sign-algorithm=sha1&q-ak=%s&q-sign-time=%s&q-key-time=%s&q-header-list=content-md5&q-url-param-list=uploadid&q-signature=%s", rspUGC.Data.TempCertificate.SecretID, keyTime, keyTime, singleKeyUpload)
	req.Header.Set("Authorization", authorization)
	req.Header.Set("X-Cos-Security-Token", token)
	req.Header.Set("Content-Md5", base64Hash)
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Host", host)
	rspFile, err := base.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer rspFile.Body.Close()
	body, err := io.ReadAll(rspFile.Body)
	if err != nil {
		return err
	}
	if strings.Contains(string(body), "Location") {
		return nil
	}
	return errors.New(string(body))
}

func (d *Vtencent) FinishUploadMaterial(SummaryKey string, VodVerifyKey string, UploadContext, VodFileId string) (RspFinishUpload, error) {
	api := "https://api.vs.tencent.com/PaaS/Material/FinishUploadMaterial"
	form := base.Json{
		"UploadContext": UploadContext,
		"VodVerifyKey":  VodVerifyKey,
		"VodFileId":     VodFileId,
		"UploadFullKey": SummaryKey}
	rsp, err := d.request(api, http.MethodPost, func(req *resty.Request) {
		req.SetBody(form)
	}, nil)
	if err != nil {
		return RspFinishUpload{}, err
	}
	var resps RspFinishUpload
	if err := json.Unmarshal(rsp, &resps); err != nil {
		return RspFinishUpload{}, err
	}
	if len(resps.Data.MaterialID) == 0 {
		return RspFinishUpload{}, errors.New(string(rsp))
	}
	return resps, nil
}

func (d *Vtencent) FileUpload(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	classId, err := strconv.Atoi(dstDir.GetID())
	if err != nil {
		return err
	}
	secretKey := "M6BrOqmpW9rl+jZuFoAOwNawGdWCaKllffpDb53za4I="
	SummaryKey := QTwoSignatureKey(uuid.New().String(), secretKey)
	rspCreatrMaterial, err := d.CreateUploadMaterial(classId, stream.GetName(), SummaryKey)
	if err != nil {
		return err
	}
	rspUGC, err := d.ApplyUploadUGC(rspCreatrMaterial.Data.VodUploadSign, stream)
	if err != nil {
		return err
	}
	uploadId, err := d.GetUploadId(rspUGC)
	if err != nil {
		return err
	}
	Etag, err := d.PutFile(uploadId, rspUGC, stream)
	if err != nil {
		return err
	}
	err = d.CompleteMultipartUpload(uploadId, Etag, rspUGC, stream)
	if err != nil {
		return err
	}
	rspCommitUGC, err := d.CommitUploadUGC(rspCreatrMaterial.Data.VodUploadSign, rspUGC.Data.VodSessionKey)
	if err != nil {
		return err
	}
	VodVerifyKey := rspCommitUGC.Data.Video.VerifyContent
	VodFileId := rspCommitUGC.Data.FileID
	UploadContext := rspCreatrMaterial.Data.UploadContext
	_, err = d.FinishUploadMaterial(SummaryKey, VodVerifyKey, UploadContext, VodFileId)
	if err != nil {
		return err
	}
	return nil
}
