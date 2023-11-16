package vtencent

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
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
		"Sort":{"Field":"UpdateTime","Order":"Desc"},
		"CreateTimeRanges":[],
		"MaterialTypes":[],
		"ReviewStatuses":[],
		"Tags":[],
		"SearchScopes":[{"Owner":{"Type":"PERSON","Id":"%s"},"ClassId":%s,"SearchOneDepth":true}]
	}`, d.TfUid, dirId)
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

func (d *Vtencent) CreateUploadMaterial(classId int, fileName string) (RspCreatrMaterial, error) {
	api := "https://api.vs.tencent.com/PaaS/Material/CreateUploadMaterial"
	form := base.Json{"Owner": base.Json{"Type": "PERSON", "Id": d.TfUid},
		"MaterialType": "IMAGE", "Name": "abklc-08cin.png", "ClassId": classId,
		"UploadSummaryKey": ""}
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

func (d *Vtencent) ChunkUpload() {

}


func (d *Vtencent) FileUpload(dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	classId, err := strconv.Atoi(dstDir.GetID())
	if err != nil {
		return err
	}
	rspCreatrMaterial, err := d.CreateUploadMaterial(classId, stream.GetName())
	if err != nil {
		return err
	}
	rspUGC, err := d.ApplyUploadUGC(rspCreatrMaterial.Data.VodUploadSign, stream)
	if err != nil {
		return err
	}
	keyTime := fmt.Sprintf("%d;%d", rspUGC.Data.Timestamp, rspUGC.Data.TempCertificate.ExpiredTime)
	singleKey := QSignatureKey(rspUGC.Data.TempCertificate.SecretKey, keyTime)
	authorization := fmt.Sprintf("q-sign-algorithm=sha1&q-ak=%s&q-sign-time=%s&q-key-time=%s&q-header-list=&q-url-param-list=prefix;uploads&q-signature=%s", rspUGC.Data.TempCertificate.SecretID, keyTime, keyTime, singleKey)
	token := rspUGC.Data.TempCertificate.Token

	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"Authorization":        authorization,
		"X-Cos-Security-Token": token,
		"X-Cos-Storage-Class":  "Standard",
	})
	StoragePath := strings.ReplaceAll(rspUGC.Data.Video.StoragePath, "/"+rspUGC.Data.StorageBucket, rspUGC.Data.StorageBucket)
	StoragePath = url.QueryEscape(StoragePath)
	api := fmt.Sprintf("https://%s-%d.cos.%s.myqcloud.com/?uploads&prefix=%s", rspUGC.Data.StorageBucket, rspUGC.Data.StorageAppID, rspUGC.Data.StorageRegionV5, StoragePath)
	res, err := req.Execute("GET", api)
	if err != nil {
		return err
	}
	res.Body()
	return nil
}
