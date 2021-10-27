package drivers

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"time"
)

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
	_,err := conf.Cache.Get(conf.Ctx,path)
	if err == nil {
		// return
	}
	
	panic("implement me")
}

func (a AliDrive) Link(path string, account *model.Account) (string, error) {
	panic("implement me")
}

func (a AliDrive) Save(account *model.Account, old *model.Account) error {
	if old != nil {
		// TODO clear something
	}
	panic("implement me")
}

var _ Driver = (*AliDrive)(nil)

func init() {
	RegisterDriver("AliDrive", &AliDrive{})
}
