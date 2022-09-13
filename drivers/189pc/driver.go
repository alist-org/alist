package _189pc

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

type Yun189PC struct {
	model.Storage
	Addition

	identity string

	client    *resty.Client
	putClient *resty.Client

	loginParam *LoginParam
	tokenInfo  *AppSessionResp
}

func (y *Yun189PC) Config() driver.Config {
	return config
}

func (y *Yun189PC) GetAddition() driver.Additional {
	return y.Addition
}

func (y *Yun189PC) Init(ctx context.Context, storage model.Storage) (err error) {
	y.Storage = storage
	if err = utils.Json.UnmarshalFromString(y.Storage.Addition, &y.Addition); err != nil {
		return err
	}

	// 处理个人云和家庭云参数
	if y.isFamily() && y.RootFolderID == "-11" {
		y.RootFolderID = ""
	}
	if !y.isFamily() && y.RootFolderID == "" {
		y.RootFolderID = "-11"
		y.FamilyID = ""
	}

	// 初始化请求客户端
	if y.client == nil {
		y.client = base.NewRestyClient().SetHeaders(map[string]string{
			"Accept":  "application/json;charset=UTF-8",
			"Referer": WEB_URL,
		})
	}
	if y.putClient == nil {
		y.putClient = base.NewRestyClient().SetTimeout(120 * time.Second)
	}

	// 避免重复登陆
	identity := utils.GetMD5Encode(y.Username + y.Password)
	if !y.isLogin() || y.identity != identity {
		y.identity = identity
		if err = y.login(); err != nil {
			return
		}
	}

	// 处理家庭云ID
	if y.isFamily() && y.FamilyID == "" {
		if y.FamilyID, err = y.getFamilyID(); err != nil {
			return err
		}
	}
	return
}

func (y *Yun189PC) Drop(ctx context.Context) error {
	return nil
}

func (y *Yun189PC) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	return y.getFiles(ctx, dir.GetID())
}

func (y *Yun189PC) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var downloadUrl struct {
		URL string `json:"fileDownloadUrl"`
	}

	fullUrl := API_URL
	if y.isFamily() {
		fullUrl += "/family/file"
	}
	fullUrl += "/getFileDownloadUrl.action"

	_, err := y.get(fullUrl, func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetQueryParam("fileId", file.GetID())
		if y.isFamily() {
			r.SetQueryParams(map[string]string{
				"familyId": y.FamilyID,
			})
		} else {
			r.SetQueryParams(map[string]string{
				"dt":   "3",
				"flag": "1",
			})
		}
	}, &downloadUrl)
	if err != nil {
		return nil, err
	}

	// 重定向获取真实链接
	downloadUrl.URL = strings.Replace(strings.ReplaceAll(downloadUrl.URL, "&amp;", "&"), "http://", "https://", 1)
	res, err := base.NoRedirectClient.R().SetContext(ctx).Get(downloadUrl.URL)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() == 302 {
		downloadUrl.URL = res.Header().Get("location")
	}

	like := &model.Link{
		URL: downloadUrl.URL,
		Header: http.Header{
			"User-Agent": []string{base.UserAgent},
		},
	}
	/*
		// 获取链接有效时常
		strs := regexp.MustCompile(`(?i)expire[^=]*=([0-9]*)`).FindStringSubmatch(downloadUrl.URL)
		if len(strs) == 2 {
			timestamp, err := strconv.ParseInt(strs[1], 10, 64)
			if err == nil {
				expired := time.Duration(timestamp-time.Now().Unix()) * time.Second
				like.Expiration = &expired
			}
		}
	*/
	return like, nil
}

func (y *Yun189PC) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	fullUrl := API_URL
	if y.isFamily() {
		fullUrl += "/family/file"
	}
	fullUrl += "/createFolder.action"

	_, err := y.post(fullUrl, func(req *resty.Request) {
		req.SetContext(ctx)
		req.SetQueryParams(map[string]string{
			"folderName":   dirName,
			"relativePath": "",
		})
		if y.isFamily() {
			req.SetQueryParams(map[string]string{
				"familyId": y.FamilyID,
				"parentId": parentDir.GetID(),
			})
		} else {
			req.SetQueryParams(map[string]string{
				"parentFolderId": parentDir.GetID(),
			})
		}
	}, nil)
	return err
}

func (y *Yun189PC) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	_, err := y.post(API_URL+"/batch/createBatchTask.action", func(req *resty.Request) {
		req.SetContext(ctx)
		req.SetFormData(map[string]string{
			"type": "MOVE",
			"taskInfos": MustString(utils.Json.MarshalToString(
				[]BatchTaskInfo{
					{
						FileId:   srcObj.GetID(),
						FileName: srcObj.GetName(),
						IsFolder: BoolToNumber(srcObj.IsDir()),
					},
				})),
			"targetFolderId": dstDir.GetID(),
		})
		if y.isFamily() {
			req.SetFormData(map[string]string{
				"familyId": y.FamilyID,
			})
		}
	}, nil)
	return err
}

func (y *Yun189PC) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	queryParam := make(map[string]string)
	fullUrl := API_URL
	method := http.MethodPost
	if y.isFamily() {
		fullUrl += "/family/file"
		method = http.MethodGet
		queryParam["familyId"] = y.FamilyID
	}
	if srcObj.IsDir() {
		fullUrl += "/renameFolder.action"
		queryParam["folderId"] = srcObj.GetID()
		queryParam["destFolderName"] = newName
	} else {
		fullUrl += "/renameFile.action"
		queryParam["fileId"] = srcObj.GetID()
		queryParam["destFileName"] = newName
	}
	_, err := y.request(fullUrl, method, func(req *resty.Request) {
		req.SetContext(ctx)
		req.SetQueryParams(queryParam)
	}, nil, nil)
	return err
}

func (y *Yun189PC) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	_, err := y.post(API_URL+"/batch/createBatchTask.action", func(req *resty.Request) {
		req.SetContext(ctx)
		req.SetFormData(map[string]string{
			"type": "COPY",
			"taskInfos": MustString(utils.Json.MarshalToString(
				[]BatchTaskInfo{
					{
						FileId:   srcObj.GetID(),
						FileName: srcObj.GetName(),
						IsFolder: BoolToNumber(srcObj.IsDir()),
					},
				})),
			"targetFolderId": dstDir.GetID(),
			"targetFileName": dstDir.GetName(),
		})
		if y.isFamily() {
			req.SetFormData(map[string]string{
				"familyId": y.FamilyID,
			})
		}
	}, nil)
	return err
}

func (y *Yun189PC) Remove(ctx context.Context, obj model.Obj) error {
	_, err := y.post(API_URL+"/batch/createBatchTask.action", func(req *resty.Request) {
		req.SetContext(ctx)
		req.SetFormData(map[string]string{
			"type": "DELETE",
			"taskInfos": MustString(utils.Json.MarshalToString(
				[]*BatchTaskInfo{
					{
						FileId:   obj.GetID(),
						FileName: obj.GetName(),
						IsFolder: BoolToNumber(obj.IsDir()),
					},
				})),
		})

		if y.isFamily() {
			req.SetFormData(map[string]string{
				"familyId": y.FamilyID,
			})
		}
	}, nil)
	return err
}

func (y *Yun189PC) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	if y.RapidUpload {
		return y.FastUpload(ctx, dstDir, stream, up)
	}
	return y.CommonUpload(ctx, dstDir, stream, up)
}
