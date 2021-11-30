package drivers

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	"path/filepath"
	"strconv"
	"time"
)

var googleClient = resty.New()

type GoogleTokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (driver GoogleDrive) RefreshToken(account *model.Account) error {
	url := "https://www.googleapis.com/oauth2/v4/token"
	var resp TokenResp
	var e GoogleTokenError
	_, err := googleClient.R().SetResult(&resp).SetError(&e).
		SetFormData(map[string]string{
			"client_id":     account.ClientId,
			"client_secret": account.ClientSecret,
			"refresh_token": account.RefreshToken,
			"grant_type":    "refresh_token",
		}).Post(url)
	if err != nil {
		return err
	}
	if e.Error != "" {
		return fmt.Errorf(e.Error)
	}
	account.AccessToken = resp.AccessToken
	account.Status = "work"
	return nil
}

type GoogleFile struct {
	Id           string     `json:"id"`
	Name         string     `json:"name"`
	MimeType     string     `json:"mimeType"`
	ModifiedTime *time.Time `json:"modifiedTime"`
	Size         string     `json:"size"`
}

func (driver GoogleDrive) IsDir(mimeType string) bool {
	return mimeType == "application/vnd.google-apps.folder" || mimeType == "application/vnd.google-apps.shortcut"
}

func (driver GoogleDrive) FormatFile(file *GoogleFile) *model.File {
	f := &model.File{
		Id:        file.Id,
		Name:      file.Name,
		Driver:    driver.Config().Name,
		UpdatedAt: file.ModifiedTime,
		Thumbnail: "",
		Url:       "",
	}
	if driver.IsDir(file.MimeType) {
		f.Type = conf.FOLDER
	} else {
		size, _ := strconv.ParseInt(file.Size, 10, 64)
		f.Size = size
		f.Type = utils.GetFileType(filepath.Ext(file.Name))
	}
	return f
}

type GoogleFiles struct {
	NextPageToken string       `json:"nextPageToken"`
	Files         []GoogleFile `json:"files"`
}

type GoogleError struct {
	Error struct {
		Errors []struct {
			Domain       string `json:"domain"`
			Reason       string `json:"reason"`
			Message      string `json:"message"`
			LocationType string `json:"location_type"`
			Location     string `json:"location"`
		}
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (driver GoogleDrive) GetFiles(id string, account *model.Account) ([]GoogleFile, error) {
	pageToken := "first"
	res := make([]GoogleFile, 0)
	for pageToken != "" {
		if pageToken == "first" {
			pageToken = ""
		}
		var resp GoogleFiles
		var e GoogleError
		_, err := googleClient.R().SetResult(&resp).SetError(&e).
			SetHeader("Authorization", "Bearer "+account.AccessToken).
			SetQueryParams(map[string]string{
				"orderBy":                   "folder,name,modifiedTime desc",
				"fields":                    "files(id,name,mimeType,size,modifiedTime),nextPageToken",
				"pageSize":                  "1000",
				"q":                         fmt.Sprintf("'%s' in parents and trashed = false", id),
				"includeItemsFromAllDrives": "true",
				"supportsAllDrives":         "true",
				"pageToken":                 pageToken,
			}).Get("https://www.googleapis.com/drive/v3/files")
		if err != nil {
			return nil, err
		}
		if e.Error.Code != 0 {
			if e.Error.Code == 401 {
				err = driver.RefreshToken(account)
				if err != nil {
					_ = model.SaveAccount(account)
					return nil, err
				}
				return driver.GetFiles(id, account)
			}
			return nil, fmt.Errorf("%s: %v", e.Error.Message, e.Error.Errors)
		}
		pageToken = resp.NextPageToken
		res = append(res, resp.Files...)
	}
	return res, nil
}

//func (driver GoogleDrive) GetFile(path string, account *model.Account) (*GoogleFile, error) {
//	dir, name := filepath.Split(path)
//	dir = utils.ParsePath(dir)
//	_, _, err := driver.Path(dir, account)
//	if err != nil {
//		return nil, err
//	}
//	parentFiles_, _ := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, dir))
//	parentFiles, _ := parentFiles_.([]GoogleFile)
//	for _, file := range parentFiles {
//		if file.Name == name {
//			if !driver.IsDir(file.MimeType) {
//				return &file, err
//			} else {
//				return nil, drivers.NotFile
//			}
//		}
//	}
//	return nil, drivers.PathNotFound
//}

func init() {
	RegisterDriver(&GoogleDrive{})
	googleClient.SetRetryCount(3)
}
