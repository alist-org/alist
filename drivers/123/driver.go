package _123

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type Pan123 struct {
	model.Storage
	Addition
	apiRateLimit sync.Map
}

func (d *Pan123) Config() driver.Config {
	return config
}

func (d *Pan123) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Pan123) Init(ctx context.Context) error {
	_, err := d.request(UserInfo, http.MethodGet, nil, nil)
	return err
}

func (d *Pan123) Drop(ctx context.Context) error {
	_, _ = d.request(Logout, http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{})
	}, nil)
	return nil
}

func (d *Pan123) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.getFiles(ctx, dir.GetID(), dir.GetName())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return src, nil
	})
}

func (d *Pan123) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	if f, ok := file.(File); ok {
		//var resp DownResp
		var headers map[string]string
		if !utils.IsLocalIPAddr(args.IP) {
			headers = map[string]string{
				//"X-Real-IP":       "1.1.1.1",
				"X-Forwarded-For": args.IP,
			}
		}
		data := base.Json{
			"driveId":   0,
			"etag":      f.Etag,
			"fileId":    f.FileId,
			"fileName":  f.FileName,
			"s3keyFlag": f.S3KeyFlag,
			"size":      f.Size,
			"type":      f.Type,
		}
		resp, err := d.request(DownloadInfo, http.MethodPost, func(req *resty.Request) {
			
			req.SetBody(data).SetHeaders(headers)
		}, nil)
		if err != nil {
			return nil, err
		}
		downloadUrl := utils.Json.Get(resp, "data", "DownloadUrl").ToString()
		u, err := url.Parse(downloadUrl)
		if err != nil {
			return nil, err
		}
		nu := u.Query().Get("params")
		if nu != "" {
			du, _ := base64.StdEncoding.DecodeString(nu)
			u, err = url.Parse(string(du))
			if err != nil {
				return nil, err
			}
		}
		u_ := u.String()
		log.Debug("download url: ", u_)
		res, err := base.NoRedirectClient.R().SetHeader("Referer", "https://www.123pan.com/").Get(u_)
		if err != nil {
			return nil, err
		}
		log.Debug(res.String())
		link := model.Link{
			URL: u_,
		}
		log.Debugln("res code: ", res.StatusCode())
		if res.StatusCode() == 302 {
			link.URL = res.Header().Get("location")
		} else if res.StatusCode() < 300 {
			link.URL = utils.Json.Get(res.Body(), "data", "redirect_url").ToString()
		}
		link.Header = http.Header{
			"Referer": []string{"https://www.123pan.com/"},
		}
		return &link, nil
	} else {
		return nil, fmt.Errorf("can't convert obj")
	}
}

func (d *Pan123) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	data := base.Json{
		"driveId":      0,
		"etag":         "",
		"fileName":     dirName,
		"parentFileId": parentDir.GetID(),
		"size":         0,
		"type":         1,
	}
	_, err := d.request(Mkdir, http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *Pan123) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	data := base.Json{
		"fileIdList":   []base.Json{{"FileId": srcObj.GetID()}},
		"parentFileId": dstDir.GetID(),
	}
	_, err := d.request(Move, http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *Pan123) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	data := base.Json{
		"driveId":  0,
		"fileId":   srcObj.GetID(),
		"fileName": newName,
	}
	_, err := d.request(Rename, http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *Pan123) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotSupport
}

func (d *Pan123) Remove(ctx context.Context, obj model.Obj) error {
	if f, ok := obj.(File); ok {
		data := base.Json{
			"driveId":           0,
			"operation":         true,
			"fileTrashInfoList": []File{f},
		}
		_, err := d.request(Trash, http.MethodPost, func(req *resty.Request) {
			req.SetBody(data)
		}, nil)
		return err
	} else {
		return fmt.Errorf("can't convert obj")
	}
}

func (d *Pan123) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// const DEFAULT int64 = 10485760
	h := md5.New()
	// need to calculate md5 of the full content
	tempFile, err := stream.CacheFullInTempFile()
	if err != nil {
		return err
	}
	defer func() {
		_ = tempFile.Close()
	}()
	if _, err = utils.CopyWithBuffer(h, tempFile); err != nil {
		return err
	}
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	etag := hex.EncodeToString(h.Sum(nil))
	data := base.Json{
		"driveId":      0,
		"duplicate":    2, // 2->覆盖 1->重命名 0->默认
		"etag":         etag,
		"fileName":     stream.GetName(),
		"parentFileId": dstDir.GetID(),
		"size":         stream.GetSize(),
		"type":         0,
	}
	var resp UploadResp
	res, err := d.request(UploadRequest, http.MethodPost, func(req *resty.Request) {
		req.SetBody(data).SetContext(ctx)
	}, &resp)
	if err != nil {
		return err
	}
	log.Debugln("upload request res: ", string(res))
	if resp.Data.Reuse || resp.Data.Key == "" {
		return nil
	}
	if resp.Data.AccessKeyId == "" || resp.Data.SecretAccessKey == "" || resp.Data.SessionToken == "" {
		err = d.newUpload(ctx, &resp, stream, tempFile, up)
		return err
	} else {
		cfg := &aws.Config{
			Credentials:      credentials.NewStaticCredentials(resp.Data.AccessKeyId, resp.Data.SecretAccessKey, resp.Data.SessionToken),
			Region:           aws.String("123pan"),
			Endpoint:         aws.String(resp.Data.EndPoint),
			S3ForcePathStyle: aws.Bool(true),
		}
		s, err := session.NewSession(cfg)
		if err != nil {
			return err
		}
		uploader := s3manager.NewUploader(s)
		if stream.GetSize() > s3manager.MaxUploadParts*s3manager.DefaultUploadPartSize {
			uploader.PartSize = stream.GetSize() / (s3manager.MaxUploadParts - 1)
		}
		input := &s3manager.UploadInput{
			Bucket: &resp.Data.Bucket,
			Key:    &resp.Data.Key,
			Body:   tempFile,
		}
		_, err = uploader.UploadWithContext(ctx, input)
	}
	_, err = d.request(UploadComplete, http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"fileId": resp.Data.FileId,
		}).SetContext(ctx)
	}, nil)
	return err
}

func (d *Pan123) APIRateLimit(ctx context.Context, api string) error {
	value, _ := d.apiRateLimit.LoadOrStore(api,
		rate.NewLimiter(rate.Every(700*time.Millisecond), 1))
	limiter := value.(*rate.Limiter)

	return limiter.Wait(ctx)
}

var _ driver.Driver = (*Pan123)(nil)
