package drivers

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strconv"
	"time"
)

type Pan123 struct {
}

var pan123Client = resty.New()

func (p Pan123) Items() []Item {
	return []Item{
		{
			Name:        "proxy",
			Label:       "proxy",
			Type:        "bool",
			Required:    true,
			Description: "allow proxy",
		},
		{
			Name:        "username",
			Label:       "username",
			Type:        "string",
			Required:    true,
			Description: "account username/phone number",
		},
		{
			Name:        "password",
			Label:       "password",
			Type:        "string",
			Required:    true,
			Description: "account password",
		},
		{
			Name:     "root_folder",
			Label:    "root folder file_id",
			Type:     "string",
			Required: false,
		},
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     "select",
			Values:   "name,fileId,updateAt,createAt",
			Required: true,
		},
		{
			Name:     "order_direction",
			Label:    "order_direction",
			Type:     "select",
			Values:   "asc,desc",
			Required: true,
		},
	}
}

type Pan123TokenResp struct {
	Code int `json:"code"`
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
	Message string `json:"message"`
}

func (p Pan123) Login(account *model.Account) error {
	var resp Pan123TokenResp
	_, err := pan123Client.R().
		SetResult(&resp).
		SetBody(Json{
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

func (p Pan123) Save(account *model.Account, old *model.Account) error {
	if account.RootFolder == "" {
		account.RootFolder = "0"
	}
	err := p.Login(account)
	return err
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

func (p Pan123) FormatFile(file *Pan123File) *model.File {
	f := &model.File{
		Name:      file.FileName,
		Size:      file.Size,
		Driver:    "123Pan",
		UpdatedAt: file.UpdateAt,
	}
	if file.Type == 1 {
		f.Type = conf.FOLDER
	} else {
		f.Type = utils.GetFileType(filepath.Ext(file.FileName))
	}
	return f
}

type Pan123Files struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		InfoList []Pan123File `json:"InfoList"`
		Next     string       `json:"Next"`
	} `json:"data"`
}

func (p Pan123) GetFiles(parentId string, account *model.Account) ([]Pan123File, error) {
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
		log.Debugf("%+v", resp)
		if resp.Code != 0 {
			if resp.Code == 401 {
				err := p.Login(account)
				if err != nil {
					return nil, err
				}
				return p.GetFiles(parentId, account)
			}
			return nil, fmt.Errorf(resp.Message)
		}
		next = resp.Data.Next
		res = append(res, resp.Data.InfoList...)
	}
	return res, nil
}

func (p Pan123) Path(path string, account *model.Account) (*model.File, []*model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("pan123 path: %s", path)
	cache, err := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path))
	if err == nil {
		files, _ := cache.([]Pan123File)
		if len(files) != 0 {
			res := make([]*model.File, 0)
			for _, file := range files {
				res = append(res, p.FormatFile(&file))
			}
			return nil, res, nil
		}
	}
	// no cache or len(files) == 0
	fileId := account.RootFolder
	if path != "/" {
		dir, name := filepath.Split(path)
		dir = utils.ParsePath(dir)
		_, _, err = p.Path(dir, account)
		if err != nil {
			return nil, nil, err
		}
		parentFiles_, _ := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, dir))
		parentFiles, _ := parentFiles_.([]Pan123File)
		found := false
		for _, file := range parentFiles {
			if file.FileName == name {
				found = true
				if file.Type != 1 {
					url, err := p.Link(path, account)
					if err != nil {
						return nil, nil, err
					}

					f := p.FormatFile(&file)
					f.Url = url
					return f, nil, nil
				} else {
					fileId = strconv.FormatInt(file.FileId, 10)
					break
				}
			}
		}
		if !found {
			return nil, nil, fmt.Errorf("path not found")
		}
	}
	files, err := p.GetFiles(fileId, account)
	if err != nil {
		return nil, nil, err
	}
	log.Debugf("%+v", files)
	_ = conf.Cache.Set(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path), files, nil)
	res := make([]*model.File, 0)
	for _, file := range files {
		res = append(res, p.FormatFile(&file))
	}
	return nil, res, nil
}

func (p Pan123) GetFile(path string, account *model.Account) (*Pan123File, error) {
	dir, name := filepath.Split(path)
	dir = utils.ParsePath(dir)
	_, _, err := p.Path(dir, account)
	if err != nil {
		return nil, err
	}
	parentFiles_, _ := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, dir))
	parentFiles, _ := parentFiles_.([]Pan123File)
	for _, file := range parentFiles {
		if file.FileName == name {
			if file.Type != 1 {
				return &file, err
			} else {
				return nil, fmt.Errorf("not file")
			}
		}
	}
	return nil, fmt.Errorf("path not found")
}

type Pan123DownResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		DownloadUrl string `json:"DownloadUrl"`
	} `json:"data"`
}

func (p Pan123) Link(path string, account *model.Account) (string, error) {
	file, err := p.GetFile(utils.ParsePath(path), account)
	if err != nil {
		return "", err
	}
	var resp Pan123DownResp
	_, err = pan123Client.R().SetResult(&resp).SetHeader("authorization", "Bearer "+account.AccessToken).
		SetBody(Json{
			"driveId":   0,
			"etag":      file.Etag,
			"fileId":    file.FileId,
			"fileName":  file.FileName,
			"s3keyFlag": file.S3KeyFlag,
			"size":      file.Size,
			"type":      file.Type,
		}).Post("https://www.123pan.com/api/file/download_info")
	if err != nil {
		return "", err
	}
	if resp.Code != 0 {
		if resp.Code == 401 {
			err := p.Login(account)
			if err != nil {
				return "", err
			}
			return p.Link(path, account)
		}
		return "", fmt.Errorf(resp.Message)
	}
	return resp.Data.DownloadUrl, nil
}

func (p Pan123) Proxy(c *gin.Context, account *model.Account) {
	c.Request.Header.Del("origin")
}

func (p Pan123) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, nil
}

var _ Driver = (*Pan123)(nil)

func init() {
	RegisterDriver("123Pan", &Pan123{})
	pan123Client.SetRetryCount(3)
}
