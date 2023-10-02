package _139

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

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
	Account string
}

func (d *Yun139) Config() driver.Config {
	return config
}

func (d *Yun139) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Yun139) Init(ctx context.Context) error {
	if d.Authorization == "" {
		return fmt.Errorf("authorization is empty")
	}
	decode, err := base64.StdEncoding.DecodeString(d.Authorization)
	if err != nil {
		return err
	}
	decodeStr := string(decode)
	splits := strings.Split(decodeStr, ":")
	if len(splits) < 2 {
		return fmt.Errorf("authorization is invalid, splits < 2")
	}
	d.Account = splits[1]
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
	_, err := d.post(pathname, data, nil)
	return err
}

func (d *Yun139) Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	if d.isFamily() {
		return nil, errs.NotImplement
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
	if err != nil {
		return nil, err
	}
	return srcObj, nil
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

const (
	_  = iota //ignore first value by assigning to blank identifier
	KB = 1 << (10 * iota)
	MB
	GB
	TB
)

func getPartSize(size int64) int64 {
	// 网盘对于分片数量存在上限
	if size/GB > 30 {
		return 512 * MB
	}
	return 100 * MB
}

func (d *Yun139) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	data := base.Json{
		"manualRename": 2,
		"operation":    0,
		"fileCount":    1,
		"totalSize":    0, // 去除上传大小限制
		"uploadContentList": []base.Json{{
			"contentName": stream.GetName(),
			"contentSize": 0, // 去除上传大小限制
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
			"totalSize":    0,
			"uploadContentList": []base.Json{{
				"contentName": stream.GetName(),
				"contentSize": 0,
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

	// Progress
	p := driver.NewProgress(stream.GetSize(), up)

	var partSize = getPartSize(stream.GetSize())
	part := (stream.GetSize() + partSize - 1) / partSize
	if part == 0 {
		part = 1
	}
	for i := int64(0); i < part; i++ {
		if utils.IsCanceled(ctx) {
			return ctx.Err()
		}

		start := i * partSize
		byteSize := stream.GetSize() - start
		if byteSize > partSize {
			byteSize = partSize
		}

		limitReader := io.LimitReader(stream, byteSize)
		// Update Progress
		r := io.TeeReader(limitReader, p)
		req, err := http.NewRequest("POST", resp.Data.UploadResult.RedirectionURL, r)
		if err != nil {
			return err
		}

		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "text/plain;name="+unicode(stream.GetName()))
		req.Header.Set("contentSize", strconv.FormatInt(stream.GetSize(), 10))
		req.Header.Set("range", fmt.Sprintf("bytes=%d-%d", start, start+byteSize-1))
		req.Header.Set("uploadtaskID", resp.Data.UploadResult.UploadTaskID)
		req.Header.Set("rangeType", "0")
		req.ContentLength = byteSize

		res, err := base.HttpClient.Do(req)
		if err != nil {
			return err
		}
		_ = res.Body.Close()
		log.Debugf("%+v", res)
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", res.StatusCode)
		}
	}

	return nil
}

var _ driver.Driver = (*Yun139)(nil)
