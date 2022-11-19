package _123

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

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
	AccessToken string
}

func (d *Pan123) Config() driver.Config {
	return config
}

func (d *Pan123) GetAddition() driver.Additional {
	return d.Addition
}

func (d *Pan123) Init(ctx context.Context, storage model.Storage) error {
	d.Storage = storage
	err := utils.Json.UnmarshalFromString(d.Storage.Addition, &d.Addition)
	if err != nil {
		return err
	}
	return d.login()
}

func (d *Pan123) Drop(ctx context.Context) error {
	return nil
}

func (d *Pan123) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.getFiles(dir.GetID())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return src, nil
	})
}

//func (d *Pan123) Get(ctx context.Context, path string) (model.Obj, error) {
//	// this is optional
//	return nil, errs.NotImplement
//}

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
		resp, err := d.request("https://www.123pan.com/api/file/download_info", http.MethodPost, func(req *resty.Request) {
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
		u_ := fmt.Sprintf("https://%s%s", u.Host, u.Path)
		res, err := base.NoRedirectClient.R().SetQueryParamsFromValues(u.Query()).Head(u_)
		if err != nil {
			return nil, err
		}
		log.Debug(res.String())
		link := model.Link{
			URL: downloadUrl,
		}
		log.Debugln("res code: ", res.StatusCode())
		if res.StatusCode() == 302 {
			link.URL = res.Header().Get("location")
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
	_, err := d.request("https://www.123pan.com/api/file/upload_request", http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *Pan123) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	data := base.Json{
		"fileIdList":   []base.Json{{"FileId": srcObj.GetID()}},
		"parentFileId": dstDir.GetID(),
	}
	_, err := d.request("https://www.123pan.com/api/file/mod_pid", http.MethodPost, func(req *resty.Request) {
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
	_, err := d.request("https://www.123pan.com/api/file/rename", http.MethodPost, func(req *resty.Request) {
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
		_, err := d.request("https://www.123pan.com/b/api/file/trash", http.MethodPost, func(req *resty.Request) {
			req.SetBody(data)
		}, nil)
		return err
	} else {
		return fmt.Errorf("can't convert obj")
	}
}

func (d *Pan123) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	const DEFAULT int64 = 10485760
	var uploadFile io.Reader
	h := md5.New()
	if d.StreamUpload && stream.GetSize() > DEFAULT {
		// 只计算前10MIB
		buf := bytes.NewBuffer(make([]byte, 0, DEFAULT))
		if n, err := io.CopyN(io.MultiWriter(buf, h), stream, DEFAULT); err != io.EOF && n == 0 {
			return err
		}
		// 增加额外参数防止MD5碰撞
		h.Write([]byte(stream.GetName()))
		num := make([]byte, 8)
		binary.BigEndian.PutUint64(num, uint64(stream.GetSize()))
		h.Write(num)
		// 拼装
		uploadFile = io.MultiReader(buf, stream)
	} else {
		// 计算完整文件MD5
		tempFile, err := utils.CreateTempFile(stream.GetReadCloser())
		if err != nil {
			return err
		}
		defer func() {
			_ = tempFile.Close()
			_ = os.Remove(tempFile.Name())
		}()
		if _, err = io.Copy(h, tempFile); err != nil {
			return err
		}
		_, err = tempFile.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
		uploadFile = tempFile
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
	_, err := d.request("https://www.123pan.com/a/api/file/upload_request", http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, &resp)
	if err != nil {
		return err
	}
	if resp.Data.Reuse || resp.Data.Key == "" {
		return nil
	}
	cfg := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(resp.Data.AccessKeyId, resp.Data.SecretAccessKey, resp.Data.SessionToken),
		Region:           aws.String("123pan"),
		Endpoint:         aws.String("file.123pan.com"),
		S3ForcePathStyle: aws.Bool(true),
	}
	s, err := session.NewSession(cfg)
	if err != nil {
		return err
	}
	uploader := s3manager.NewUploader(s)
	input := &s3manager.UploadInput{
		Bucket: &resp.Data.Bucket,
		Key:    &resp.Data.Key,
		Body:   uploadFile,
	}
	_, err = uploader.Upload(input)
	if err != nil {
		return err
	}
	_, err = d.request("https://www.123pan.com/api/file/upload_complete", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"fileId": resp.Data.FileId,
		})
	}, nil)
	return err
}

var _ driver.Driver = (*Pan123)(nil)
