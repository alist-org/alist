package _139

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type Yun139 struct {
	model.Storage
	Addition
}

func (d *Yun139) Config() driver.Config {
	return config
}

func (d *Yun139) GetAddition() driver.Additional {
	return d.Addition
}

func (d *Yun139) Init(ctx context.Context, storage model.Storage) error {
	d.Storage = storage
	err := utils.Json.UnmarshalFromString(d.Storage.Addition, &d.Addition)
	if err != nil {
		return err
	}
	_, err = d.post("/orchestration/personalCloud/user/v1.0/qryUserExternInfo", base.Json{
		"qryUserExternInfoReq": base.Json{
			"commonAccountInfo": base.Json{
				"account":     d.Account,
				"accountType": 1,
			},
		},
	}, nil)
	return err
}

func (d *Yun139) Drop(ctx context.Context) error {
	return nil
}

func (d *Yun139) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	if d.isFamily() {
		return d.familyGetFiles(dir.GetID())
	} else {
		return d.getFiles(dir.GetID())
	}
}

//func (d *Yun139) Get(ctx context.Context, path string) (model.Obj, error) {
//	// this is optional
//	return nil, errs.NotImplement
//}

func (d *Yun139) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	u, err := d.getLink(file.GetID())
	if err != nil {
		return nil, err
	}
	return &model.Link{URL: u}, nil
}

func (d *Yun139) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	data := base.Json{
		"createCatalogExtReq": base.Json{
			"parentCatalogID": parentDir.GetID(),
			"newCatalogName":  dirName,
			"commonAccountInfo": base.Json{
				"account":     d.Account,
				"accountType": 1,
			},
		},
	}
	pathname := "/orchestration/personalCloud/catalog/v1.0/createCatalogExt"
	if d.isFamily() {
		data = base.Json{
			"cloudID": d.CloudID,
			"commonAccountInfo": base.Json{
				"account":     d.Account,
				"accountType": 1,
			},
			"docLibName": dirName,
		}
		pathname = "/orchestration/familyCloud/cloudCatalog/v1.0/createCloudDoc"
	}
	_, err := d.post(pathname,
		data, nil)
	return err
}

func (d *Yun139) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	if d.isFamily() {
		return errs.NotImplement
	}
	var contentInfoList []string
	var catalogInfoList []string
	if srcObj.IsDir() {
		catalogInfoList = append(catalogInfoList, srcObj.GetID())
	} else {
		contentInfoList = append(contentInfoList, srcObj.GetID())
	}
	data := base.Json{
		"createBatchOprTaskReq": base.Json{
			"taskType":   3,
			"actionType": "304",
			"taskInfo": base.Json{
				"contentInfoList": contentInfoList,
				"catalogInfoList": catalogInfoList,
				"newCatalogID":    dstDir.GetID(),
			},
			"commonAccountInfo": base.Json{
				"account":     d.Account,
				"accountType": 1,
			},
		},
	}
	pathname := "/orchestration/personalCloud/batchOprTask/v1.0/createBatchOprTask"
	_, err := d.post(pathname, data, nil)
	return err
}

func (d *Yun139) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	if d.isFamily() {
		return errs.NotImplement
	}
	var data base.Json
	var pathname string
	if srcObj.IsDir() {
		data = base.Json{
			"catalogID":   srcObj.GetID(),
			"catalogName": newName,
			"commonAccountInfo": base.Json{
				"account":     d.Account,
				"accountType": 1,
			},
		}
		pathname = "/orchestration/personalCloud/catalog/v1.0/updateCatalogInfo"
	} else {
		data = base.Json{
			"contentID":   srcObj.GetID(),
			"contentName": newName,
			"commonAccountInfo": base.Json{
				"account":     d.Account,
				"accountType": 1,
			},
		}
		pathname = "/orchestration/personalCloud/content/v1.0/updateContentInfo"
	}
	_, err := d.post(pathname, data, nil)
	return err
}

func (d *Yun139) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	if d.isFamily() {
		return errs.NotImplement
	}
	var contentInfoList []string
	var catalogInfoList []string
	if srcObj.IsDir() {
		catalogInfoList = append(catalogInfoList, srcObj.GetID())
	} else {
		contentInfoList = append(contentInfoList, srcObj.GetID())
	}
	data := base.Json{
		"createBatchOprTaskReq": base.Json{
			"taskType":   3,
			"actionType": 309,
			"taskInfo": base.Json{
				"contentInfoList": contentInfoList,
				"catalogInfoList": catalogInfoList,
				"newCatalogID":    dstDir.GetID(),
			},
			"commonAccountInfo": base.Json{
				"account":     d.Account,
				"accountType": 1,
			},
		},
	}
	pathname := "/orchestration/personalCloud/batchOprTask/v1.0/createBatchOprTask"
	_, err := d.post(pathname, data, nil)
	return err
}

func (d *Yun139) Remove(ctx context.Context, obj model.Obj) error {
	var contentInfoList []string
	var catalogInfoList []string
	if obj.IsDir() {
		catalogInfoList = append(catalogInfoList, obj.GetID())
	} else {
		contentInfoList = append(contentInfoList, obj.GetID())
	}
	data := base.Json{
		"createBatchOprTaskReq": base.Json{
			"taskType":   2,
			"actionType": 201,
			"taskInfo": base.Json{
				"newCatalogID":    "",
				"contentInfoList": contentInfoList,
				"catalogInfoList": catalogInfoList,
			},
			"commonAccountInfo": base.Json{
				"account":     d.Account,
				"accountType": 1,
			},
		},
	}
	pathname := "/orchestration/personalCloud/batchOprTask/v1.0/createBatchOprTask"
	if d.isFamily() {
		data = base.Json{
			"catalogList": catalogInfoList,
			"contentList": contentInfoList,
			"commonAccountInfo": base.Json{
				"account":     d.Account,
				"accountType": 1,
			},
			"sourceCatalogType": 1002,
			"taskType":          2,
		}
		pathname = "/orchestration/familyCloud/batchOprTask/v1.0/createBatchOprTask"
	}
	_, err := d.post(pathname, data, nil)
	return err
}

func (d *Yun139) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	data := base.Json{
		"manualRename": 2,
		"operation":    0,
		"fileCount":    1,
		"totalSize":    stream.GetSize(),
		"uploadContentList": []base.Json{{
			"contentName": stream.GetName(),
			"contentSize": stream.GetSize(),
			// "digest": "5a3231986ce7a6b46e408612d385bafa"
		}},
		"parentCatalogID": dstDir.GetID(),
		"newCatalogName":  "",
		"commonAccountInfo": base.Json{
			"account":     d.Account,
			"accountType": 1,
		},
	}
	pathname := "/orchestration/personalCloud/uploadAndDownload/v1.0/pcUploadFileRequest"
	if d.isFamily() {
		data = d.newJson(base.Json{
			"fileCount":    1,
			"manualRename": 2,
			"operation":    0,
			"path":         "",
			"seqNo":        "",
			"totalSize":    stream.GetSize(),
			"uploadContentList": []base.Json{{
				"contentName": stream.GetName(),
				"contentSize": stream.GetSize(),
				// "digest": "5a3231986ce7a6b46e408612d385bafa"
			}},
		})
		pathname = "/orchestration/familyCloud/content/v1.0/getFileUploadURL"
		return errs.NotImplement
	}
	var resp UploadResp
	_, err := d.post(pathname, data, &resp)
	if err != nil {
		return err
	}
	var Default int64 = 10485760
	part := int(math.Ceil(float64(stream.GetSize()) / float64(Default)))
	var start int64 = 0
	for i := 0; i < part; i++ {
		byteSize := stream.GetSize() - start
		if byteSize > Default {
			byteSize = Default
		}
		byteData := make([]byte, byteSize)
		_, err = io.ReadFull(stream, byteData)
		if err != nil {
			return err
		}
		req, err := http.NewRequest("POST", resp.Data.UploadResult.RedirectionURL, bytes.NewBuffer(byteData))
		if err != nil {
			return err
		}
		headers := map[string]string{
			"Accept":         "*/*",
			"Content-Type":   "text/plain;name=" + unicode(stream.GetName()),
			"contentSize":    strconv.FormatInt(stream.GetSize(), 10),
			"range":          fmt.Sprintf("bytes=%d-%d", start, start+byteSize-1),
			"content-length": strconv.FormatInt(byteSize, 10),
			"uploadtaskID":   resp.Data.UploadResult.UploadTaskID,
			"rangeType":      "0",
			"Referer":        "https://yun.139.com/",
			"User-Agent":     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36 Edg/95.0.1020.44",
			"x-SvcType":      "1",
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		res, err := base.HttpClient.Do(req)
		if err != nil {
			return err
		}
		log.Debugf("%+v", res)
		res.Body.Close()
		start += byteSize
		up(i * 100 / part)
	}
	return nil
}

var _ driver.Driver = (*Yun139)(nil)
