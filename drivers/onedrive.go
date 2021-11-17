package drivers

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

type Onedrive struct{}

var oneClient = resty.New()

type OnedriveHost struct {
	Oauth string
	Api   string
}

var onedriveHostMap = map[string]OnedriveHost{
	"global": {
		Oauth: "https://login.microsoftonline.com",
		Api:   "https://graph.microsoft.com",
	},
	"cn": {
		Oauth: "https://login.chinacloudapi.cn",
		Api:   "https://microsoftgraph.chinacloudapi.cn",
	},
	"us": {
		Oauth: "https://login.microsoftonline.us",
		Api:   "https://graph.microsoft.us",
	},
	"de": {
		Oauth: "https://login.microsoftonline.de",
		Api:   "https://graph.microsoft.de",
	},
}

func init() {
	RegisterDriver("Onedrive", &Onedrive{})
	oneClient.SetRetryCount(3)
}

func (o Onedrive) GetMetaUrl(account *model.Account, auth bool, path string) string {
	path = filepath.Join(account.RootFolder, path)
	log.Debugf(path)
	host, _ := onedriveHostMap[account.Zone]
	if auth {
		return host.Oauth
	}
	switch account.OnedriveType {
	case "onedrive":
		{
			if path == "/" || path == "\\" {
				return fmt.Sprintf("%s/v1.0/me/drive/root", host.Api)
			} else {
				return fmt.Sprintf("%s/v1.0/me/drive/root:%s:", host.Api, path)
			}
		}
	case "sharepoint":
		{
			if path == "/" {
				return fmt.Sprintf("%s/v1.0/sites/%s/drive/root", host.Api, account.SiteId)
			} else {
				return fmt.Sprintf("%s/v1.0/sites/%s/drive/root:%s:", host.Api, account.SiteId, path)
			}
		}
	default:
		return ""
	}
}

func (o Onedrive) Items() []Item {
	return []Item{
		{
			Name:        "zone",
			Label:       "zone",
			Type:        "select",
			Required:    true,
			Values:      "global,cn,us,de",
			Description: "",
		},
		{
			Name:     "onedrive_type",
			Label:    "onedrive type",
			Type:     "select",
			Required: true,
			Values:   "onedrive,sharepoint",
		},
		{
			Name:     "client_id",
			Label:    "client id",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "client_secret",
			Label:    "client secret",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "redirect_uri",
			Label:    "redirect uri",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "refresh_token",
			Label:    "refresh token",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "site_id",
			Label:    "site id",
			Type:     "string",
			Required: false,
		},
		{
			Name:     "root_folder",
			Label:    "root folder path",
			Type:     "string",
			Required: false,
		},
	}
}

type OneTokenErr struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (o Onedrive) RefreshToken(account *model.Account) error {
	url := o.GetMetaUrl(account, true, "") + "/common/oauth2/v2.0/token"
	var resp TokenResp
	var e OneTokenErr
	_, err := oneClient.R().SetResult(&resp).SetError(&e).SetFormData(map[string]string{
		"grant_type":    "refresh_token",
		"client_id":     account.ClientId,
		"client_secret": account.ClientSecret,
		"redirect_uri":  account.RedirectUri,
		"refresh_token": account.RefreshToken,
	}).Post(url)
	if err != nil {
		account.Status = err.Error()
		return err
	}
	if e.Error != "" {
		account.Status = e.ErrorDescription
		return fmt.Errorf("%s", e.ErrorDescription)
	}else {
		account.Status = "work"
	}
	account.RefreshToken, account.AccessToken = resp.RefreshToken, resp.AccessToken
	return nil
}

type OneFile struct {
	Name                 string     `json:"name"`
	Size                 int64      `json:"size"`
	LastModifiedDateTime *time.Time `json:"lastModifiedDateTime"`
	Url                  string     `json:"@microsoft.graph.downloadUrl"`
	File                 struct {
		MimeType string `json:"mimeType"`
	} `json:"file"`
}

type OneFiles struct {
	Value []OneFile `json:"value"`
}

type OneRespErr struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (o Onedrive) FormatFile(file *OneFile) *model.File {
	f := &model.File{
		Name:      file.Name,
		Size:      file.Size,
		UpdatedAt: file.LastModifiedDateTime,
		Driver:    "OneDrive",
		Url:       file.Url,
	}
	if file.File.MimeType == "" {
		f.Type = conf.FOLDER
	} else {
		f.Type = utils.GetFileType(filepath.Ext(file.Name))
	}
	return f
}

func (o Onedrive) GetFiles(account *model.Account, path string) ([]OneFile, error) {
	var files OneFiles
	var e OneRespErr
	_, err := oneClient.R().SetResult(&files).SetError(&e).
		SetHeader("Authorization", "Bearer  "+account.AccessToken).
		Get(o.GetMetaUrl(account, false, path) + "/children")
	if err != nil {
		return nil, err
	}
	if e.Error.Code != "" {
		return nil, fmt.Errorf("%s", e.Error.Message)
	}
	return files.Value, nil
}

func (o Onedrive) GetFile(account *model.Account, path string) (*OneFile, error) {
	var file OneFile
	var e OneRespErr
	_, err := oneClient.R().SetResult(&file).SetError(&e).
		SetHeader("Authorization", "Bearer  "+account.AccessToken).
		Get(o.GetMetaUrl(account, false, path))
	if err != nil {
		return nil, err
	}
	if e.Error.Code != "" {
		return nil, fmt.Errorf("%s", e.Error.Message)
	}
	return &file, nil
}

func (o Onedrive) Path(path string, account *model.Account) (*model.File, []*model.File, error) {
	path = utils.ParsePath(path)
	cache, err := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path))
	if err == nil {
		files, _ := cache.([]*model.File)
		return nil, files, nil
	}
	file, err := o.GetFile(account, path)
	if err != nil {
		return nil, nil, err
	}
	if file.File.MimeType != "" {
		return o.FormatFile(file), nil, nil
	} else {
		files, err := o.GetFiles(account, path)
		if err != nil {
			return nil, nil, err
		}
		res := make([]*model.File, 0)
		for _, file := range files {
			res = append(res, o.FormatFile(&file))
		}
		_ = conf.Cache.Set(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path), res, nil)
		return nil, res, nil
	}
}

func (o Onedrive) Link(path string, account *model.Account) (string, error) {
	file, err := o.GetFile(account, path)
	if err != nil {
		return "", err
	}
	if file.File.MimeType == "" {
		return "", fmt.Errorf("can't down folder")
	}
	return file.Url, nil
}

func (o Onedrive) Save(account *model.Account, old *model.Account) error {
	_, ok := onedriveHostMap[account.Zone]
	if !ok {
		return fmt.Errorf("no [%s] zone", account.Zone)
	}
	if old != nil {
		conf.Cron.Remove(cron.EntryID(old.CronId))
	}
	account.RootFolder = utils.ParsePath(account.RootFolder)
	err := o.RefreshToken(account)
	if err != nil {
		return err
	}
	cronId, err := conf.Cron.AddFunc("@every 1h", func() {
		name := account.Name
		log.Debugf("onedrive account name: %s", name)
		newAccount, ok := model.GetAccount(name)
		log.Debugf("onedrive account: %+v", newAccount)
		if !ok {
			return
		}
		err = o.RefreshToken(&newAccount)
		_ = model.SaveAccount(&newAccount)
	})
	if err != nil {
		return err
	}
	account.CronId = int(cronId)
	err = model.SaveAccount(account)
	if err != nil {
		return err
	}
	return nil
}

func (o Onedrive) Proxy(c *gin.Context) {
	c.Request.Header.Del("Origin")
}

func (o Onedrive) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, nil
}

var _ Driver = (*Onedrive)(nil)
