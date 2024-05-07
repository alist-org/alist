package _189pc

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/pkg/utils"
)

// 居然有四种返回方式
type RespErr struct {
	ResCode    any    `json:"res_code"` // int or string
	ResMessage string `json:"res_message"`

	Error_ string `json:"error"`

	XMLName xml.Name `xml:"error"`
	Code    string   `json:"code" xml:"code"`
	Message string   `json:"message" xml:"message"`
	Msg     string   `json:"msg"`

	ErrorCode string `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}

func (e *RespErr) HasError() bool {
	switch v := e.ResCode.(type) {
	case int, int64, int32:
		return v != 0
	case string:
		return e.ResCode != ""
	}
	return (e.Code != "" && e.Code != "SUCCESS") || e.ErrorCode != "" || e.Error_ != ""
}

func (e *RespErr) Error() string {
	switch v := e.ResCode.(type) {
	case int, int64, int32:
		if v != 0 {
			return fmt.Sprintf("res_code: %d ,res_msg: %s", v, e.ResMessage)
		}
	case string:
		if e.ResCode != "" {
			return fmt.Sprintf("res_code: %s ,res_msg: %s", e.ResCode, e.ResMessage)
		}
	}

	if e.Code != "" && e.Code != "SUCCESS" {
		if e.Msg != "" {
			return fmt.Sprintf("code: %s ,msg: %s", e.Code, e.Msg)
		}
		if e.Message != "" {
			return fmt.Sprintf("code: %s ,msg: %s", e.Code, e.Message)
		}
		return "code: " + e.Code
	}

	if e.ErrorCode != "" {
		return fmt.Sprintf("err_code: %s ,err_msg: %s", e.ErrorCode, e.ErrorMsg)
	}

	if e.Error_ != "" {
		return fmt.Sprintf("error: %s ,message: %s", e.ErrorCode, e.Message)
	}
	return ""
}

// 登陆需要的参数
type LoginParam struct {
	// 加密后的用户名和密码
	RsaUsername string
	RsaPassword string

	// rsa密钥
	jRsaKey string

	// 请求头参数
	Lt    string
	ReqId string

	// 表单参数
	ParamId string

	// 验证码
	CaptchaToken string
}

// 登陆加密相关
type EncryptConfResp struct {
	Result int `json:"result"`
	Data   struct {
		UpSmsOn   string `json:"upSmsOn"`
		Pre       string `json:"pre"`
		PreDomain string `json:"preDomain"`
		PubKey    string `json:"pubKey"`
	} `json:"data"`
}

type LoginResp struct {
	Msg    string `json:"msg"`
	Result int    `json:"result"`
	ToUrl  string `json:"toUrl"`
}

// 刷新session返回
type UserSessionResp struct {
	ResCode    int    `json:"res_code"`
	ResMessage string `json:"res_message"`

	LoginName string `json:"loginName"`

	KeepAlive       int `json:"keepAlive"`
	GetFileDiffSpan int `json:"getFileDiffSpan"`
	GetUserInfoSpan int `json:"getUserInfoSpan"`

	// 个人云
	SessionKey    string `json:"sessionKey"`
	SessionSecret string `json:"sessionSecret"`
	// 家庭云
	FamilySessionKey    string `json:"familySessionKey"`
	FamilySessionSecret string `json:"familySessionSecret"`
}

// 登录返回
type AppSessionResp struct {
	UserSessionResp

	IsSaveName string `json:"isSaveName"`

	// 会话刷新Token
	AccessToken string `json:"accessToken"`
	//Token刷新
	RefreshToken string `json:"refreshToken"`
}

// 家庭云账户
type FamilyInfoListResp struct {
	FamilyInfoResp []FamilyInfoResp `json:"familyInfoResp"`
}
type FamilyInfoResp struct {
	Count      int    `json:"count"`
	CreateTime string `json:"createTime"`
	FamilyID   int64  `json:"familyId"`
	RemarkName string `json:"remarkName"`
	Type       int    `json:"type"`
	UseFlag    int    `json:"useFlag"`
	UserRole   int    `json:"userRole"`
}

/*文件部分*/
// 文件
type Cloud189File struct {
	ID   String `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Md5  string `json:"md5"`

	LastOpTime Time `json:"lastOpTime"`
	CreateDate Time `json:"createDate"`
	Icon       struct {
		//iconOption 5
		SmallUrl string `json:"smallUrl"`
		LargeUrl string `json:"largeUrl"`

		// iconOption 10
		Max600    string `json:"max600"`
		MediumURL string `json:"mediumUrl"`
	} `json:"icon"`

	// Orientation int64  `json:"orientation"`
	// FileCata   int64  `json:"fileCata"`
	// MediaType   int    `json:"mediaType"`
	// Rev         string `json:"rev"`
	// StarLabel   int64  `json:"starLabel"`
}

func (c *Cloud189File) CreateTime() time.Time {
	return time.Time(c.CreateDate)
}

func (c *Cloud189File) GetHash() utils.HashInfo {
	return utils.NewHashInfo(utils.MD5, c.Md5)
}

func (c *Cloud189File) GetSize() int64     { return c.Size }
func (c *Cloud189File) GetName() string    { return c.Name }
func (c *Cloud189File) ModTime() time.Time { return time.Time(c.LastOpTime) }
func (c *Cloud189File) IsDir() bool        { return false }
func (c *Cloud189File) GetID() string      { return string(c.ID) }
func (c *Cloud189File) GetPath() string    { return "" }
func (c *Cloud189File) Thumb() string      { return c.Icon.SmallUrl }

// 文件夹
type Cloud189Folder struct {
	ID       String `json:"id"`
	ParentID int64  `json:"parentId"`
	Name     string `json:"name"`

	LastOpTime Time `json:"lastOpTime"`
	CreateDate Time `json:"createDate"`

	// FileListSize int64 `json:"fileListSize"`
	// FileCount int64 `json:"fileCount"`
	// FileCata  int64 `json:"fileCata"`
	// Rev          string `json:"rev"`
	// StarLabel    int64  `json:"starLabel"`
}

func (c *Cloud189Folder) CreateTime() time.Time {
	return time.Time(c.CreateDate)
}

func (c *Cloud189Folder) GetHash() utils.HashInfo {
	return utils.HashInfo{}
}

func (c *Cloud189Folder) GetSize() int64     { return 0 }
func (c *Cloud189Folder) GetName() string    { return c.Name }
func (c *Cloud189Folder) ModTime() time.Time { return time.Time(c.LastOpTime) }
func (c *Cloud189Folder) IsDir() bool        { return true }
func (c *Cloud189Folder) GetID() string      { return string(c.ID) }
func (c *Cloud189Folder) GetPath() string    { return "" }

type Cloud189FilesResp struct {
	//ResCode    int    `json:"res_code"`
	//ResMessage string `json:"res_message"`
	FileListAO struct {
		Count      int              `json:"count"`
		FileList   []Cloud189File   `json:"fileList"`
		FolderList []Cloud189Folder `json:"folderList"`
	} `json:"fileListAO"`
}

// TaskInfo 任务信息
type BatchTaskInfo struct {
	// FileId 文件ID
	FileId string `json:"fileId"`
	// FileName 文件名
	FileName string `json:"fileName"`
	// IsFolder 是否是文件夹，0-否，1-是
	IsFolder int `json:"isFolder"`
	// SrcParentId 文件所在父目录ID
	SrcParentId string `json:"srcParentId,omitempty"`

	/* 冲突管理 */
	// 1 -> 跳过 2 -> 保留 3 -> 覆盖
	DealWay    int `json:"dealWay,omitempty"`
	IsConflict int `json:"isConflict,omitempty"`
}

/* 上传部分 */
type InitMultiUploadResp struct {
	//Code string `json:"code"`
	Data struct {
		UploadType     int    `json:"uploadType"`
		UploadHost     string `json:"uploadHost"`
		UploadFileID   string `json:"uploadFileId"`
		FileDataExists int    `json:"fileDataExists"`
	} `json:"data"`
}
type UploadUrlsResp struct {
	Code string                    `json:"code"`
	Data map[string]UploadUrlsData `json:"uploadUrls"`
}
type UploadUrlsData struct {
	RequestURL    string `json:"requestURL"`
	RequestHeader string `json:"requestHeader"`
}

type UploadUrlInfo struct {
	PartNumber int
	Headers    map[string]string
	UploadUrlsData
}

type UploadProgress struct {
	UploadInfo  InitMultiUploadResp
	UploadParts []string
}

/* 第二种上传方式 */
type CreateUploadFileResp struct {
	// 上传文件请求ID
	UploadFileId int64 `json:"uploadFileId"`
	// 上传文件数据的URL路径
	FileUploadUrl string `json:"fileUploadUrl"`
	// 上传文件完成后确认路径
	FileCommitUrl string `json:"fileCommitUrl"`
	// 文件是否已存在云盘中，0-未存在，1-已存在
	FileDataExists int `json:"fileDataExists"`
}

type GetUploadFileStatusResp struct {
	CreateUploadFileResp

	// 已上传的大小
	DataSize int64 `json:"dataSize"`
	Size     int64 `json:"size"`
}

func (r *GetUploadFileStatusResp) GetSize() int64 {
	return r.DataSize + r.Size
}

type CommitMultiUploadFileResp struct {
	File struct {
		UserFileID String `json:"userFileId"`
		FileName   string `json:"fileName"`
		FileSize   int64  `json:"fileSize"`
		FileMd5    string `json:"fileMd5"`
		CreateDate Time   `json:"createDate"`
	} `json:"file"`
}

func (f *CommitMultiUploadFileResp) toFile() *Cloud189File {
	return &Cloud189File{
		ID:         f.File.UserFileID,
		Name:       f.File.FileName,
		Size:       f.File.FileSize,
		Md5:        f.File.FileMd5,
		LastOpTime: f.File.CreateDate,
		CreateDate: f.File.CreateDate,
	}
}

type OldCommitUploadFileResp struct {
	XMLName    xml.Name `xml:"file"`
	ID         String   `xml:"id"`
	Name       string   `xml:"name"`
	Size       int64    `xml:"size"`
	Md5        string   `xml:"md5"`
	CreateDate Time     `xml:"createDate"`
}

func (f *OldCommitUploadFileResp) toFile() *Cloud189File {
	return &Cloud189File{
		ID:         f.ID,
		Name:       f.Name,
		Size:       f.Size,
		Md5:        f.Md5,
		CreateDate: f.CreateDate,
		LastOpTime: f.CreateDate,
	}
}

type CreateBatchTaskResp struct {
	TaskID string `json:"taskId"`
}

type BatchTaskStateResp struct {
	FailedCount         int     `json:"failedCount"`
	Process             int     `json:"process"`
	SkipCount           int     `json:"skipCount"`
	SubTaskCount        int     `json:"subTaskCount"`
	SuccessedCount      int     `json:"successedCount"`
	SuccessedFileIDList []int64 `json:"successedFileIdList"`
	TaskID              string  `json:"taskId"`
	TaskStatus          int     `json:"taskStatus"` //1 初始化 2 存在冲突 3 执行中，4 完成
}

type BatchTaskConflictTaskInfoResp struct {
	SessionKey     string `json:"sessionKey"`
	TargetFolderID int    `json:"targetFolderId"`
	TaskID         string `json:"taskId"`
	TaskInfos      []BatchTaskInfo
	TaskType       int `json:"taskType"`
}

/* query 加密参数*/
type Params map[string]string

func (p Params) Set(k, v string) {
	p[k] = v
}

func (p Params) Encode() string {
	if p == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i := range keys {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(keys[i])
		buf.WriteByte('=')
		buf.WriteString(p[keys[i]])
	}
	return buf.String()
}
