package drivers

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/go-resty/resty/v2"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"time"
)

var aliClient = resty.New()

func init() {
	aliClient.
		SetRetryCount(3).
		SetHeader("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36").
		SetHeader("content-type", "application/json").
		SetHeader("origin", "https://aliyundrive.com")
}

type AliDrive struct {
}

type AliRespError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type AliFile struct {
	AliRespError
	DriveId            string                 `json:"drive_id"`
	CreatedAt          *time.Time             `json:"created_at"`
	DomainId           string                 `json:"domain_id"`
	EncryptMode        string                 `json:"encrypt_mode"`
	FileExtension      string                 `json:"file_extension"`
	FileId             string                 `json:"file_id"`
	Hidden             bool                   `json:"hidden"`
	Name               string                 `json:"name"`
	ParentFileId       string                 `json:"parent_file_id"`
	Starred            bool                   `json:"starred"`
	Status             string                 `json:"status"`
	Type               string                 `json:"type"`
	UpdatedAt          *time.Time             `json:"updated_at"`
	Category           string                 `json:"category"`
	ContentHash        string                 `json:"content_hash"`
	ContentHashName    string                 `json:"content_hash_name"`
	ContentType        string                 `json:"content_type"`
	Crc64Hash          string                 `json:"crc_64_hash"`
	DownloadUrl        string                 `json:"download_url"`
	PunishFlag         int64                  `json:"punish_flag"`
	Size               int64                  `json:"size"`
	Thumbnail          string                 `json:"thumbnail"`
	Url                string                 `json:"url"`
	ImageMediaMetadata map[string]interface{} `json:"image_media_metadata"`
}

func (a AliDrive) Path(path string, account *model.Account) (*model.File, []*model.File, error) {
	_, err := conf.Cache.Get(conf.Ctx, path)
	if err == nil {
		// return
	}

	panic("implement me")
}

func (a AliDrive) Link(path string, account *model.Account) (string, error) {
	panic("implement me")
}

type AliTokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func AliRefreshToken(refresh string) (string, string, error) {
	url := "https://auth.aliyundrive.com/v2/account/token"
	var resp AliTokenResp
	var e AliRespError
	_, err := aliClient.R().
		//ForceContentType("application/json").
		SetBody(JsonStr(Json{"refresh_token": refresh, "grant_type": "refresh_token"})).
		SetResult(&resp).
		SetError(&e).
		Post(url)
	if err != nil {
		return "", "", err
	}
	log.Debugf("%+v,%+v", resp, e)
	if e.Code != "" {
		return "", "", fmt.Errorf("failed to refresh token: %s", e.Message)
	}
	return resp.RefreshToken, resp.AccessToken, nil
}

func (a AliDrive) Save(account *model.Account, old *model.Account) error {
	if old != nil {
		conf.Cron.Remove(cron.EntryID(old.CronId))
	}
	refresh, access, err := AliRefreshToken(account.RefreshToken)
	if err != nil {
		return err
	}
	account.RefreshToken, account.AccessToken = refresh, access
	cronId, err := conf.Cron.AddFunc("@every 2h", func() {
		name := account.Name
		newAccount, ok := model.GetAccount(name)
		if !ok {
			return
		}
		newAccount.RefreshToken, newAccount.AccessToken, err = AliRefreshToken(newAccount.RefreshToken)
		if err != nil {
			newAccount.Status = err.Error()
		}
		_ = model.SaveAccount(newAccount)
	})
	if err != nil {
		return err
	}
	account.CronId = int(cronId)
	err = model.SaveAccount(*account)
	if err != nil {
		return err
	}
	return nil
}

var _ Driver = (*AliDrive)(nil)

func init() {
	RegisterDriver("AliDrive", &AliDrive{})
}
