package thunder_browser

import (
	"fmt"
	"strconv"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	hash_extend "github.com/alist-org/alist/v3/pkg/utils/hash"
)

type ErrResp struct {
	ErrorCode        int64  `json:"error_code"`
	ErrorMsg         string `json:"error"`
	ErrorDescription string `json:"error_description"`
	//	ErrorDetails   interface{} `json:"error_details"`
}

func (e *ErrResp) IsError() bool {
	return e.ErrorCode != 0 || e.ErrorMsg != "" || e.ErrorDescription != ""
}

func (e *ErrResp) Error() string {
	return fmt.Sprintf("ErrorCode: %d ,Error: %s ,ErrorDescription: %s ", e.ErrorCode, e.ErrorMsg, e.ErrorDescription)
}

/*
* 验证码Token
**/
type CaptchaTokenRequest struct {
	Action       string            `json:"action"`
	CaptchaToken string            `json:"captcha_token"`
	ClientID     string            `json:"client_id"`
	DeviceID     string            `json:"device_id"`
	Meta         map[string]string `json:"meta"`
	RedirectUri  string            `json:"redirect_uri"`
}

type CaptchaTokenResponse struct {
	CaptchaToken string `json:"captcha_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Url          string `json:"url"`
}

/*
* 登录
**/
type TokenResp struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`

	Sub    string `json:"sub"`
	UserID string `json:"user_id"`

	Token string `json:"token"` // "超级保险箱" 访问Token
}

func (t *TokenResp) GetToken() string {
	return fmt.Sprint(t.TokenType, " ", t.AccessToken)
}

// GetSpaceToken 获取"超级保险箱" 访问Token
func (t *TokenResp) GetSpaceToken() string {
	return t.Token
}

type SignInRequest struct {
	CaptchaToken string `json:"captcha_token"`

	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`

	Username string `json:"username"`
	Password string `json:"password"`
}

/*
* 文件
**/
type FileList struct {
	Kind            string  `json:"kind"`
	NextPageToken   string  `json:"next_page_token"`
	Files           []Files `json:"files"`
	Version         string  `json:"version"`
	VersionOutdated bool    `json:"version_outdated"`
	FolderType      int8
}

type Link struct {
	URL    string    `json:"url"`
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
	Type   string    `json:"type"`
}

var _ model.Obj = (*Files)(nil)

type Files struct {
	Kind     string `json:"kind"`
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
	Name     string `json:"name"`
	//UserID         string    `json:"user_id"`
	Size string `json:"size"`
	//Revision       string    `json:"revision"`
	//FileExtension  string    `json:"file_extension"`
	//MimeType       string    `json:"mime_type"`
	//Starred        bool      `json:"starred"`
	WebContentLink string     `json:"web_content_link"`
	CreatedTime    CustomTime `json:"created_time"`
	ModifiedTime   CustomTime `json:"modified_time"`
	IconLink       string     `json:"icon_link"`
	ThumbnailLink  string     `json:"thumbnail_link"`
	Md5Checksum    string     `json:"md5_checksum"`
	Hash           string     `json:"hash"`
	// Links map[string]Link `json:"links"`
	// Phase string          `json:"phase"`
	// Audit struct {
	// 	Status  string `json:"status"`
	// 	Message string `json:"message"`
	// 	Title   string `json:"title"`
	// } `json:"audit"`
	Medias []struct {
		//Category       string `json:"category"`
		//IconLink       string `json:"icon_link"`
		//IsDefault      bool   `json:"is_default"`
		//IsOrigin       bool   `json:"is_origin"`
		//IsVisible      bool   `json:"is_visible"`
		Link Link `json:"link"`
		//MediaID        string `json:"media_id"`
		//MediaName      string `json:"media_name"`
		//NeedMoreQuota  bool   `json:"need_more_quota"`
		//Priority       int    `json:"priority"`
		//RedirectLink   string `json:"redirect_link"`
		//ResolutionName string `json:"resolution_name"`
		// Video          struct {
		// 	AudioCodec string `json:"audio_codec"`
		// 	BitRate    int    `json:"bit_rate"`
		// 	Duration   int    `json:"duration"`
		// 	FrameRate  int    `json:"frame_rate"`
		// 	Height     int    `json:"height"`
		// 	VideoCodec string `json:"video_codec"`
		// 	VideoType  string `json:"video_type"`
		// 	Width      int    `json:"width"`
		// } `json:"video"`
		// VipTypes []string `json:"vip_types"`
	} `json:"medias"`
	Trashed     bool   `json:"trashed"`
	DeleteTime  string `json:"delete_time"`
	OriginalURL string `json:"original_url"`
	//Params            struct{} `json:"params"`
	//OriginalFileIndex int    `json:"original_file_index"`
	Space string `json:"space"`
	//Apps              []interface{} `json:"apps"`
	//Writable   bool   `json:"writable"`
	FolderType string `json:"folder_type"`
	//Collection interface{} `json:"collection"`
	SortName         string     `json:"sort_name"`
	UserModifiedTime CustomTime `json:"user_modified_time"`
	//SpellName         []interface{} `json:"spell_name"`
	//FileCategory      string        `json:"file_category"`
	//Tags              []interface{} `json:"tags"`
	//ReferenceEvents   []interface{} `json:"reference_events"`
	//ReferenceResource interface{}   `json:"reference_resource"`
	//Params0           struct {
	//	PlatformIcon   string `json:"platform_icon"`
	//	SmallThumbnail string `json:"small_thumbnail"`
	//} `json:"params,omitempty"`
}

func (c *Files) GetHash() utils.HashInfo {
	return utils.NewHashInfo(hash_extend.GCID, c.Hash)
}

func (c *Files) GetSize() int64        { size, _ := strconv.ParseInt(c.Size, 10, 64); return size }
func (c *Files) GetName() string       { return c.Name }
func (c *Files) CreateTime() time.Time { return c.CreatedTime.Time }
func (c *Files) ModTime() time.Time    { return c.ModifiedTime.Time }
func (c *Files) IsDir() bool           { return c.Kind == FOLDER }
func (c *Files) GetID() string         { return c.ID }
func (c *Files) GetPath() string {
	return ""
}
func (c *Files) Thumb() string { return c.ThumbnailLink }

func (c *Files) GetSpace() string {
	if c.Space != "" {
		return c.Space
	} else {
		// "迅雷云盘" 文件夹内 Space 为空
		return ""
	}
}

/*
* 上传
**/
type UploadTaskResponse struct {
	UploadType string `json:"upload_type"`

	/*//UPLOAD_TYPE_FORM
	Form struct {
		//Headers struct{} `json:"headers"`
		Kind       string `json:"kind"`
		Method     string `json:"method"`
		MultiParts struct {
			OSSAccessKeyID string `json:"OSSAccessKeyId"`
			Signature      string `json:"Signature"`
			Callback       string `json:"callback"`
			Key            string `json:"key"`
			Policy         string `json:"policy"`
			XUserData      string `json:"x:user_data"`
		} `json:"multi_parts"`
		URL string `json:"url"`
	} `json:"form"`*/

	//UPLOAD_TYPE_RESUMABLE
	Resumable struct {
		Kind   string `json:"kind"`
		Params struct {
			AccessKeyID     string    `json:"access_key_id"`
			AccessKeySecret string    `json:"access_key_secret"`
			Bucket          string    `json:"bucket"`
			Endpoint        string    `json:"endpoint"`
			Expiration      time.Time `json:"expiration"`
			Key             string    `json:"key"`
			SecurityToken   string    `json:"security_token"`
		} `json:"params"`
		Provider string `json:"provider"`
	} `json:"resumable"`

	File Files `json:"file"`
}
