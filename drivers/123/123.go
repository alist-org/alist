package _23

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"math/rand"
	"path/filepath"
	"strconv"
	"time"
)

var pan123Client = resty.New()

type Pan123TokenResp struct {
	Code int `json:"code"`
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
	Message string `json:"message"`
}

type Pan123File struct {
	FileName  string     `json:"FileName"`
	Size      int64      `json:"Size"`
	UpdateAt  *time.Time `json:"UpdateAt"`
	FileId    int64      `json:"FileId"`
	Type      int        `json:"Type"`
	Etag      string     `json:"Etag"`
	S3KeyFlag string     `json:"S3KeyFlag"`
}

type Pan123Files struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		InfoList []Pan123File `json:"InfoList"`
		Next     string       `json:"Next"`
	} `json:"data"`
}

type Pan123DownResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		DownloadUrl string `json:"DownloadUrl"`
	} `json:"data"`
}

func (driver Pan123) Login(account *model.Account) error {
	var resp Pan123TokenResp
	_, err := pan123Client.R().
		SetResult(&resp).
		SetBody(base.Json{
			"passport": account.Username,
			"password": account.Password,
		}).Post("https://www.123pan.com/api/user/sign_in")
	if err != nil {
		return err
	}
	if resp.Code != 200 {
		err = fmt.Errorf(resp.Message)
		account.Status = resp.Message
	} else {
		account.Status = "work"
		account.AccessToken = resp.Data.Token
	}
	_ = model.SaveAccount(account)
	return err
}

func (driver Pan123) FormatFile(file *Pan123File) *model.File {
	f := &model.File{
		Id:        strconv.FormatInt(file.FileId, 10),
		Name:      file.FileName,
		Size:      file.Size,
		Driver:    driver.Config().Name,
		UpdatedAt: file.UpdateAt,
	}
	if file.Type == 1 {
		f.Type = conf.FOLDER
	} else {
		f.Type = utils.GetFileType(filepath.Ext(file.FileName))
	}
	return f
}

func (driver Pan123) GetFiles(parentId string, account *model.Account) ([]Pan123File, error) {
	next := "0"
	res := make([]Pan123File, 0)
	for next != "-1" {
		var resp Pan123Files
		_, err := pan123Client.R().SetResult(&resp).
			SetHeader("authorization", "Bearer "+account.AccessToken).
			SetQueryParams(map[string]string{
				"driveId":        "0",
				"limit":          "100",
				"next":           next,
				"orderBy":        account.OrderBy,
				"orderDirection": account.OrderDirection,
				"parentFileId":   parentId,
				"trashed":        "false",
			}).Get("https://www.123pan.com/api/file/list")
		if err != nil {
			return nil, err
		}
		if resp.Code != 0 {
			if resp.Code == 401 {
				err := driver.Login(account)
				if err != nil {
					return nil, err
				}
				return driver.GetFiles(parentId, account)
			}
			return nil, fmt.Errorf(resp.Message)
		}
		next = resp.Data.Next
		res = append(res, resp.Data.InfoList...)
	}
	return res, nil
}

func (driver Pan123) Post(url string, data base.Json, account *model.Account) ([]byte, error) {
	res, err := pan123Client.R().
		SetHeader("authorization", "Bearer "+account.AccessToken).
		SetBody(data).Post(url)
	if err != nil {
		return nil, err
	}
	body := res.Body()
	if jsoniter.Get(body, "code").ToInt() != 0 {
		return nil, errors.New(jsoniter.Get(body, "message").ToString())
	}
	return body, nil
}

func (driver Pan123) GetFile(path string, account *model.Account) (*Pan123File, error) {
	dir, name := filepath.Split(path)
	dir = utils.ParsePath(dir)
	_, err := driver.Files(dir, account)
	if err != nil {
		return nil, err
	}
	parentFiles_, _ := base.GetCache(dir, account)
	parentFiles, _ := parentFiles_.([]Pan123File)
	for _, file := range parentFiles {
		if file.FileName == name {
			if file.Type != conf.FOLDER {
				return &file, err
			} else {
				return nil, base.ErrNotFile
			}
		}
	}
	return nil, base.ErrPathNotFound
}

func RandStr(length int) string {
	str := "123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	var result []byte
	rand.Seed(time.Now().UnixNano()+ int64(rand.Intn(100)))
	for i := 0; i < length; i++ {
		result = append(result, bytes[rand.Intn(len(bytes))])
	}
	return string(result)
}

func HMAC(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	//	fmt.Println(h.Sum(nil))
	//sha := hex.EncodeToString(h.Sum(nil))
	//	fmt.Println(sha)
	//return sha
	return string(h.Sum(nil))
}

func init() {
	base.RegisterDriver(&Pan123{})
	pan123Client.SetRetryCount(3)
}
