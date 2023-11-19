package vtencent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-resty/resty/v2"
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
	var resps RspFiles
	rsp, err := d.request(api, http.MethodPost, func(req *resty.Request) {
		req.SetBody(form).ForceContentType("application/json")
	}, &resps)
	if err != nil {
	}
	if err := json.Unmarshal(rsp, &resps); err != nil {
		return []File{}, err
	}
	return resps.Data.ResourceInfoSet, nil
}

func (d *Vtencent) CreateUploadMaterial(classId int, fileName string, UploadSummaryKey string) (RspCreatrMaterial, error) {
	api := "https://api.vs.tencent.com/PaaS/Material/CreateUploadMaterial"
	form := base.Json{"Owner": base.Json{"Type": "PERSON", "Id": d.TfUid},
		"MaterialType": "VIDEO", "Name": fileName, "ClassId": classId,
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

func (d *Vtencent) FinishHashUploadMaterial(SummaryKey string, UploadContext string) (RspFinishUpload, error) {
	api := "https://api.vs.tencent.com/PaaS/Material/FinishUploadMaterial"
	form := base.Json{
		"UploadContext": UploadContext,
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
	tmpF, err := stream.CacheFullInTempFile()
	if err != nil {
		return err
	}
	SummaryKey, err := utils.HashFile(utils.SHA1, tmpF)
	if err != nil {
		return err
	}
	chunkHash := SummaryKey
	var chunkLength int64 = 1024 * 1024 * 10
	if stream.GetSize() > chunkLength {
		buf := bytes.NewBuffer(make([]byte, 0, chunkLength))
		io.CopyN(buf, tmpF, chunkLength)
		chunkHash = utils.HashData(utils.SHA1, buf.Bytes())
	}
	rspCreatrMaterial, err := d.CreateUploadMaterial(classId, stream.GetName(), chunkHash)
	if err != nil {
		return err
	}
	if rspCreatrMaterial.Data.QuickUpload {
		UploadContext := rspCreatrMaterial.Data.UploadContext
		_, err = d.FinishHashUploadMaterial(SummaryKey, UploadContext)
		if err != nil {
			return err
		}
		return nil
	}
	rspUGC, err := d.ApplyUploadUGC(rspCreatrMaterial.Data.VodUploadSign, stream)
	if err != nil {
		return err
	}
	params := rspUGC.Data
	certificate := params.TempCertificate
	cfg := &aws.Config{
		HTTPClient: base.HttpClient,
		// S3ForcePathStyle: aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(certificate.SecretID, certificate.SecretKey, certificate.Token),
		Region:      aws.String(params.StorageRegionV5),
		Endpoint:    aws.String(fmt.Sprintf("cos.%s.myqcloud.com", params.StorageRegionV5)),
	}
	ss, err := session.NewSession(cfg)
	if err != nil {
		return err
	}
	uploader := s3manager.NewUploader(ss)
	input := &s3manager.UploadInput{
		Bucket: aws.String(fmt.Sprintf("%s-%d", params.StorageBucket, params.StorageAppID)),
		Key:    &params.Video.StoragePath,
		Body:   stream,
	}
	_, err = uploader.UploadWithContext(ctx, input)
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
