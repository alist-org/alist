package xunlei

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type XunLeiCloud struct{}

func init() {
	base.RegisterDriver(new(XunLeiCloud))
}

func (driver XunLeiCloud) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:      "XunLeiCloud",
		LocalSort: true,
	}
}

func (driver XunLeiCloud) Items() []base.Item {
	return []base.Item{
		{
			Name:        "username",
			Label:       "username",
			Type:        base.TypeString,
			Required:    true,
			Description: "account username/phone number",
		},
		{
			Name:        "password",
			Label:       "password",
			Type:        base.TypeString,
			Required:    true,
			Description: "account password",
		},
		{
			Name:  "captcha_token",
			Label: "verified captcha token",
			Type:  base.TypeString,
		},
		{
			Name:  "root_folder",
			Label: "root folder file_id",
			Type:  base.TypeString,
		},
		{
			Name:     "client_version",
			Label:    "client version",
			Default:  "7.43.0.7998",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "client_id",
			Label:    "client id",
			Default:  "Xp6vsxz_7IYVw2BB",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "client_secret",
			Label:    "client secret",
			Default:  "Xp6vsy4tN9toTVdMSpomVdXpRmES",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "algorithms",
			Label:    "algorithms",
			Default:  "hrVPGbeqYPs+CIscj05VpAtjalzY5yjpvlMS8bEo,DrI0uTP,HHK0VXyMgY0xk2K0o,BBaXsExvL3GadmIacjWv7ISUJp3ifAwqbJumu,5toJ7ejB+bh1,5LsZTFAFjgvFvIl1URBgOAJ,QcJ5Ry+,hYgZVz8r7REROaCYfd9,zw6gXgkk/8TtGrmx6EGfekPESLnbZfDFwqR,gtSwLnMBa8h12nF3DU6+LwEQPHxd,fMG8TvtAYbCkxuEbIm0Xi/Lb7Z",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "package_name",
			Label:    "package name",
			Default:  "com.xunlei.downloadprovider",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "user_agent",
			Label:    "user agent",
			Default:  "ANDROID-com.xunlei.downloadprovider/7.43.0.7998 netWorkType/WIFI appid/40 deviceName/Samsung_Sm-g9810 deviceModel/SM-G9810 OSVersion/7.1.2 protocolVersion/301 platformVersion/10 sdkVersion/220200 Oauth2Client/0.9 (Linux 4_0_9+) (JAVA 0)",
			Type:     base.TypeString,
			Required: false,
		},
		{
			Name:     "device_id",
			Label:    "device id",
			Default:  utils.GetMD5Encode(uuid.NewString()),
			Type:     base.TypeString,
			Required: true,
		},
	}
}

func (driver XunLeiCloud) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}

	client := GetClient(account)
	// 指定验证通过的captchaToken
	if account.CaptchaToken != "" {
		client.UpdateCaptchaToken(strings.TrimSpace(account.CaptchaToken))
		account.CaptchaToken = ""
	}

	if client.token == "" {
		return client.Login(account)
	}

	account.Status = "work"
	model.SaveAccount(account)
	return nil
}

func (driver XunLeiCloud) File(path string, account *model.Account) (*model.File, error) {
	path = utils.ParsePath(path)
	if path == "/" {
		return &model.File{
			Id:        account.RootFolder,
			Name:      account.Name,
			Size:      0,
			Type:      conf.FOLDER,
			Driver:    driver.Config().Name,
			UpdatedAt: account.UpdatedAt,
		}, nil
	}
	dir, name := filepath.Split(path)
	files, err := driver.Files(dir, account)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.Name == name {
			return &file, nil
		}
	}
	return nil, base.ErrPathNotFound
}

func (driver XunLeiCloud) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ := cache.([]model.File)
		return files, nil
	}

	parentFile, err := driver.File(path, account)
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Millisecond * 300)
	files := make([]model.File, 0)
	var pageToken string
	for {
		var fileList FileList
		_, err = GetClient(account).Request("GET", FILE_API_URL, func(r *resty.Request) {
			r.SetQueryParams(map[string]string{
				"parent_id":  parentFile.Id,
				"page_token": pageToken,
				"with_audit": "true",
				"limit":      "100",
				"filters":    `{"phase": {"eq": "PHASE_TYPE_COMPLETE"}, "trashed":{"eq":false}}`,
			})
			r.SetResult(&fileList)
		}, account)
		if err != nil {
			return nil, err
		}
		for _, file := range fileList.Files {
			if file.Kind == FOLDER || (file.Kind == FILE && file.Audit.Status == "STATUS_OK") {
				files = append(files, *driver.formatFile(&file))
			}
		}
		if fileList.NextPageToken == "" {
			break
		}
		pageToken = fileList.NextPageToken
	}
	if len(files) > 0 {
		_ = base.SetCache(path, files, account)
	}
	return files, nil
}

func (driver XunLeiCloud) formatFile(file *Files) *model.File {
	size, _ := strconv.ParseInt(file.Size, 10, 64)
	tp := conf.FOLDER
	if file.Kind == FILE {
		tp = utils.GetFileType(file.FileExtension)
	}
	return &model.File{
		Id:        file.ID,
		Name:      file.Name,
		Size:      size,
		Type:      tp,
		Driver:    driver.Config().Name,
		UpdatedAt: file.CreatedTime,
		Thumbnail: file.ThumbnailLink,
	}
}

func (driver XunLeiCloud) Link(args base.Args, account *model.Account) (*base.Link, error) {
	file, err := driver.File(utils.ParsePath(args.Path), account)
	if err != nil {
		return nil, err
	}
	if file.Type == conf.FOLDER {
		return nil, base.ErrNotFile
	}
	var lFile Files
	clinet := GetClient(account)
	_, err = clinet.Request("GET", FILE_API_URL+"/{fileID}", func(r *resty.Request) {
		r.SetPathParam("fileID", file.Id)
		r.SetQueryParam("with_audit", "true")
		r.SetResult(&lFile)
	}, account)
	if err != nil {
		return nil, err
	}
	return &base.Link{
		Headers: []base.Header{
			{Name: "User-Agent", Value: clinet.userAgent},
		},
		Url: lFile.WebContentLink,
	}, nil
}

func (driver XunLeiCloud) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	file, err := driver.File(path, account)
	if err != nil {
		return nil, nil, err
	}
	if !file.IsDir() {
		return file, nil, nil
	}
	files, err := driver.Files(path, account)
	if err != nil {
		return nil, nil, err
	}
	return nil, files, nil
}

func (driver XunLeiCloud) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver XunLeiCloud) Rename(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	_, err = GetClient(account).Request("PATCH", FILE_API_URL+"/{fileID}", func(r *resty.Request) {
		r.SetPathParam("fileID", srcFile.Id)
		r.SetBody(&base.Json{"name": filepath.Base(dst)})
	}, account)
	return err
}

func (driver XunLeiCloud) MakeDir(path string, account *model.Account) error {
	dir, name := filepath.Split(path)
	parentFile, err := driver.File(dir, account)
	if err != nil {
		return err
	}
	if !parentFile.IsDir() {
		return base.ErrNotFolder
	}
	_, err = GetClient(account).Request("POST", FILE_API_URL, func(r *resty.Request) {
		r.SetBody(&base.Json{
			"kind":      FOLDER,
			"name":      name,
			"parent_id": parentFile.Id,
		})
	}, account)
	return err
}

func (driver XunLeiCloud) Move(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}

	dstDirFile, err := driver.File(filepath.Dir(dst), account)
	if err != nil {
		return err
	}

	_, err = GetClient(account).Request("POST", FILE_API_URL+":batchMove", func(r *resty.Request) {
		r.SetBody(&base.Json{
			"to":  base.Json{"parent_id": dstDirFile.Id},
			"ids": []string{srcFile.Id},
		})
	}, account)
	return err
}

func (driver XunLeiCloud) Copy(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}

	dstDirFile, err := driver.File(filepath.Dir(dst), account)
	if err != nil {
		return err
	}
	_, err = GetClient(account).Request("POST", FILE_API_URL+":batchCopy", func(r *resty.Request) {
		r.SetBody(&base.Json{
			"to":  base.Json{"parent_id": dstDirFile.Id},
			"ids": []string{srcFile.Id},
		})
	}, account)
	return err
}

func (driver XunLeiCloud) Delete(path string, account *model.Account) error {
	srcFile, err := driver.File(path, account)
	if err != nil {
		return err
	}
	_, err = GetClient(account).Request("PATCH", FILE_API_URL+"/{fileID}/trash", func(r *resty.Request) {
		r.SetPathParam("fileID", srcFile.Id)
		r.SetBody(&base.Json{})
	}, account)
	return err
}

func (driver XunLeiCloud) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}

	parentFile, err := driver.File(file.ParentPath, account)
	if err != nil {
		return err
	}

	/*
		tempFile, err := ioutil.TempFile(conf.Conf.TempDir, "file-*")
		if err != nil {
			return err
		}

		defer func() {
			_ = tempFile.Close()
			_ = os.Remove(tempFile.Name())
		}()

		gcid, err := getGcid(io.TeeReader(file, tempFile), int64(file.Size))
		if err != nil {
			return err
		}

		if _, err = tempFile.Seek(0, io.SeekStart); err != nil {
			return err
		}
	*/

	var resp UploadTaskResponse
	_, err = GetClient(account).Request("POST", FILE_API_URL, func(r *resty.Request) {
		r.SetBody(&base.Json{
			"kind":        FILE,
			"parent_id":   parentFile.Id,
			"name":        file.Name,
			"size":        file.Size,
			"hash":        "1CF254FBC456E1B012CD45C546636AA62CF8350E",
			"upload_type": UPLOAD_TYPE_RESUMABLE,
		})
		r.SetResult(&resp)
	}, account)
	if err != nil {
		return err
	}

	param := resp.Resumable.Params
	if resp.UploadType == UPLOAD_TYPE_RESUMABLE {
		param.Endpoint = strings.TrimLeft(param.Endpoint, param.Bucket+".")
		s, err := session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(param.AccessKeyID, param.AccessKeySecret, param.SecurityToken),
			Region:      aws.String("xunlei"),
			Endpoint:    aws.String(param.Endpoint),
		})
		if err != nil {
			return err
		}
		_, err = s3manager.NewUploader(s).Upload(&s3manager.UploadInput{
			Bucket:  aws.String(param.Bucket),
			Key:     aws.String(param.Key),
			Expires: aws.Time(param.Expiration),
			Body:    file,
		})
		return err
	}
	return nil
}

var _ base.Driver = (*XunLeiCloud)(nil)
