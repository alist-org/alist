package pikpak

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alist-org/alist/v3/internal/op"
	"net/http"
	"strconv"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	hash_extend "github.com/alist-org/alist/v3/pkg/utils/hash"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type PikPak struct {
	model.Storage
	Addition
	*Common
	oauth2Token oauth2.TokenSource
}

func (d *PikPak) Config() driver.Config {
	return config
}

func (d *PikPak) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *PikPak) Init(ctx context.Context) (err error) {
	if d.ClientID == "" || d.ClientSecret == "" {
		d.ClientID = "YNxT9w7GMdWvEOKa"
		d.ClientSecret = "dbw2OtmVEeuUvIptb1Coyg"
	}

	if d.Common == nil {
		d.Common = &Common{
			client:       base.NewRestyClient(),
			CaptchaToken: "",
			UserID:       "",
			DeviceID:     utils.GetMD5EncodeStr(d.Username + d.Password),
			UserAgent:    BuildCustomUserAgent(utils.GetMD5EncodeStr(d.Username+d.Password), ClientID, PackageName, SdkVersion, ClientVersion, PackageName, ""),
			RefreshCTokenCk: func(token string) {
				d.Common.CaptchaToken = token
				op.MustSaveDriverStorage(d)
			},
		}
	}

	oauth2Config := &oauth2.Config{
		ClientID:     d.ClientID,
		ClientSecret: d.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://user.mypikpak.com/v1/auth/signin",
			TokenURL:  "https://user.mypikpak.com/v1/auth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	d.oauth2Token = oauth2.ReuseTokenSource(nil, utils.TokenSource(func() (*oauth2.Token, error) {
		return oauth2Config.PasswordCredentialsToken(
			context.WithValue(context.Background(), oauth2.HTTPClient, base.HttpClient),
			d.Username,
			d.Password,
		)
	}))

	// 获取CaptchaToken
	_ = d.RefreshCaptchaTokenAtLogin(GetAction(http.MethodGet, "https://api-drive.mypikpak.com/drive/v1/files"), d.Username)

	// 获取用户ID
	_ = d.GetUserID(ctx)
	// 更新UserAgent
	d.Common.UserAgent = BuildCustomUserAgent(d.Common.DeviceID, ClientID, PackageName, SdkVersion, ClientVersion, PackageName, d.Common.UserID)
	return nil
}

func (d *PikPak) Drop(ctx context.Context) error {
	return nil
}

func (d *PikPak) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.getFiles(dir.GetID())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return fileToObj(src), nil
	})
}

func (d *PikPak) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var resp File
	_, err := d.requestWithCaptchaToken(fmt.Sprintf("https://api-drive.mypikpak.com/drive/v1/files/%s?_magic=2021&thumbnail_size=SIZE_LARGE", file.GetID()),
		http.MethodGet, nil, &resp)
	if err != nil {
		return nil, err
	}
	link := model.Link{
		URL: resp.WebContentLink,
	}
	if !d.DisableMediaLink && len(resp.Medias) > 0 && resp.Medias[0].Link.Url != "" {
		log.Debugln("use media link")
		link.URL = resp.Medias[0].Link.Url
	}
	return &link, nil
}

func (d *PikPak) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"kind":      "drive#folder",
			"parent_id": parentDir.GetID(),
			"name":      dirName,
		})
	}, nil)
	return err
}

func (d *PikPak) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files:batchMove", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"ids": []string{srcObj.GetID()},
			"to": base.Json{
				"parent_id": dstDir.GetID(),
			},
		})
	}, nil)
	return err
}

func (d *PikPak) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files/"+srcObj.GetID(), http.MethodPatch, func(req *resty.Request) {
		req.SetBody(base.Json{
			"name": newName,
		})
	}, nil)
	return err
}

func (d *PikPak) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files:batchCopy", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"ids": []string{srcObj.GetID()},
			"to": base.Json{
				"parent_id": dstDir.GetID(),
			},
		})
	}, nil)
	return err
}

func (d *PikPak) Remove(ctx context.Context, obj model.Obj) error {
	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files:batchTrash", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"ids": []string{obj.GetID()},
		})
	}, nil)
	return err
}

func (d *PikPak) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	hi := stream.GetHash()
	sha1Str := hi.GetHash(hash_extend.GCID)
	if len(sha1Str) < hash_extend.GCID.Width {
		tFile, err := stream.CacheFullInTempFile()
		if err != nil {
			return err
		}

		sha1Str, err = utils.HashFile(hash_extend.GCID, tFile, stream.GetSize())
		if err != nil {
			return err
		}
	}

	var resp UploadTaskData
	res, err := d.request("https://api-drive.mypikpak.com/drive/v1/files", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"kind":        "drive#file",
			"name":        stream.GetName(),
			"size":        stream.GetSize(),
			"hash":        strings.ToUpper(sha1Str),
			"upload_type": "UPLOAD_TYPE_RESUMABLE",
			"objProvider": base.Json{"provider": "UPLOAD_TYPE_UNKNOWN"},
			"parent_id":   dstDir.GetID(),
			"folder_type": "NORMAL",
		})
	}, &resp)
	if err != nil {
		return err
	}

	// 秒传成功
	if resp.Resumable == nil {
		log.Debugln(string(res))
		return nil
	}

	params := resp.Resumable.Params
	endpoint := strings.Join(strings.Split(params.Endpoint, ".")[1:], ".")
	cfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(params.AccessKeyID, params.AccessKeySecret, params.SecurityToken),
		Region:      aws.String("pikpak"),
		Endpoint:    &endpoint,
	}
	ss, err := session.NewSession(cfg)
	if err != nil {
		return err
	}
	uploader := s3manager.NewUploader(ss)
	if stream.GetSize() > s3manager.MaxUploadParts*s3manager.DefaultUploadPartSize {
		uploader.PartSize = stream.GetSize() / (s3manager.MaxUploadParts - 1)
	}
	input := &s3manager.UploadInput{
		Bucket: &params.Bucket,
		Key:    &params.Key,
		Body:   stream,
	}
	_, err = uploader.UploadWithContext(ctx, input)
	return err
}

// 离线下载文件
func (d *PikPak) OfflineDownload(ctx context.Context, fileUrl string, parentDir model.Obj, fileName string) (*OfflineTask, error) {
	requestBody := base.Json{
		"kind":        "drive#file",
		"name":        fileName,
		"upload_type": "UPLOAD_TYPE_URL",
		"url": base.Json{
			"url": fileUrl,
		},
		"parent_id":   parentDir.GetID(),
		"folder_type": "",
	}

	var resp OfflineDownloadResp
	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files", http.MethodPost, func(req *resty.Request) {
		req.SetBody(requestBody)
	}, &resp)

	if err != nil {
		return nil, err
	}

	return &resp.Task, err
}

/*
获取离线下载任务列表
phase 可能的取值：
PHASE_TYPE_RUNNING, PHASE_TYPE_ERROR, PHASE_TYPE_COMPLETE, PHASE_TYPE_PENDING
*/
func (d *PikPak) OfflineList(ctx context.Context, nextPageToken string, phase []string) ([]OfflineTask, error) {
	res := make([]OfflineTask, 0)
	url := "https://api-drive.mypikpak.com/drive/v1/tasks"

	if len(phase) == 0 {
		phase = []string{"PHASE_TYPE_RUNNING", "PHASE_TYPE_ERROR", "PHASE_TYPE_COMPLETE", "PHASE_TYPE_PENDING"}
	}
	params := map[string]string{
		"type":           "offline",
		"thumbnail_size": "SIZE_SMALL",
		"limit":          "10000",
		"page_token":     nextPageToken,
		"with":           "reference_resource",
	}

	// 处理 phase 参数
	if len(phase) > 0 {
		filters := base.Json{
			"phase": map[string]string{
				"in": strings.Join(phase, ","),
			},
		}
		filtersJSON, err := json.Marshal(filters)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal filters: %w", err)
		}
		params["filters"] = string(filtersJSON)
	}

	var resp OfflineListResp
	_, err := d.request(url, http.MethodGet, func(req *resty.Request) {
		req.SetContext(ctx).
			SetQueryParams(params)
	}, &resp)

	if err != nil {
		return nil, fmt.Errorf("failed to get offline list: %w", err)
	}
	res = append(res, resp.Tasks...)
	return res, nil
}

func (d *PikPak) DeleteOfflineTasks(ctx context.Context, taskIDs []string, deleteFiles bool) error {
	url := "https://api-drive.mypikpak.com/drive/v1/tasks"
	params := map[string]string{
		"task_ids":     strings.Join(taskIDs, ","),
		"delete_files": strconv.FormatBool(deleteFiles),
	}
	_, err := d.request(url, http.MethodDelete, func(req *resty.Request) {
		req.SetContext(ctx).
			SetQueryParams(params)
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to delete tasks %v: %w", taskIDs, err)
	}
	return nil
}

func (d *PikPak) GetUserID(ctx context.Context) error {
	url := "https://api-drive.mypikpak.com/vip/v1/vip/info"
	var resp VipInfo
	_, err := d.requestWithCaptchaToken(url, http.MethodGet, func(req *resty.Request) {
		req.SetContext(ctx)
	}, &resp)
	if err != nil {
		return fmt.Errorf("failed to get user id : %w", err)
	}
	if resp.Data.UserID != "" {
		d.Common.SetUserID(resp.Data.UserID)
	}
	return nil
}

var _ driver.Driver = (*PikPak)(nil)
