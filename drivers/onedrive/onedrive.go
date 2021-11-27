package onedrive

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

var oneClient = resty.New()

type Host struct {
	Oauth string
	Api   string
}

var onedriveHostMap = map[string]Host{
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

func (driver Onedrive) GetMetaUrl(account *model.Account, auth bool, path string) string {
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

type OneTokenErr struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (driver Onedrive) RefreshToken(account *model.Account) error {
	url := driver.GetMetaUrl(account, true, "") + "/common/oauth2/v2.0/token"
	var resp drivers.TokenResp
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
	} else {
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
	Value    []OneFile `json:"value"`
	NextLink string    `json:"@odata.nextLink"`
}

type OneRespErr struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (driver Onedrive) FormatFile(file *OneFile) *model.File {
	f := &model.File{
		Name:      file.Name,
		Size:      file.Size,
		UpdatedAt: file.LastModifiedDateTime,
		Driver:    driverName,
		Url:       file.Url,
	}
	if file.File.MimeType == "" {
		f.Type = conf.FOLDER
	} else {
		f.Type = utils.GetFileType(filepath.Ext(file.Name))
	}
	return f
}

func (driver Onedrive) GetFiles(account *model.Account, path string) ([]OneFile, error) {
	var res []OneFile
	nextLink := driver.GetMetaUrl(account, false, path) + "/children"
	if account.OrderBy != "" {
		nextLink += fmt.Sprintf("?orderby=%s", account.OrderBy)
		if account.OrderDirection != "" {
			nextLink += fmt.Sprintf(" %s", account.OrderDirection)
		}
	}
	for nextLink != "" {
		var files OneFiles
		var e OneRespErr
		_, err := oneClient.R().SetResult(&files).SetError(&e).
			SetHeader("Authorization", "Bearer  "+account.AccessToken).
			Get(nextLink)
		if err != nil {
			return nil, err
		}
		if e.Error.Code != "" {
			return nil, fmt.Errorf("%s", e.Error.Message)
		}
		res = append(res, files.Value...)
		nextLink = files.NextLink
	}
	return res, nil
}

func (driver Onedrive) GetFile(account *model.Account, path string) (*OneFile, error) {
	var file OneFile
	var e OneRespErr
	_, err := oneClient.R().SetResult(&file).SetError(&e).
		SetHeader("Authorization", "Bearer  "+account.AccessToken).
		Get(driver.GetMetaUrl(account, false, path))
	if err != nil {
		return nil, err
	}
	if e.Error.Code != "" {
		return nil, fmt.Errorf("%s", e.Error.Message)
	}
	return &file, nil
}

var _ drivers.Driver = (*Onedrive)(nil)

func init() {
	drivers.RegisterDriver(driverName, &Onedrive{})
	oneClient.SetRetryCount(3)
}