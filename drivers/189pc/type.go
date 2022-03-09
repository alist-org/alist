package _189

import "encoding/xml"

type LoginParam struct {
	CaptchaToken string
	Lt           string
	ParamId      string
	ReqId        string
	jRsaKey      string

	vCodeID string
	vCodeRS string
}

// 居然有四种返回方式
type Erron struct {
	ResCode    string `json:"res_code"`
	ResMessage string `json:"res_message"`

	XMLName xml.Name `xml:"error"`
	Code    string   `json:"code" xml:"code"`
	Message string   `json:"message" xml:"message"`

	// Code    string `json:"code"`
	Msg string `json:"msg"`

	ErrorCode string `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
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

//登录返回
type appSessionResp struct {
	UserSessionResp

	IsSaveName string `json:"isSaveName"`

	// 会话刷新Token
	AccessToken string `json:"accessToken"`
	//Token刷新
	RefreshToken string `json:"refreshToken"`
}

/*文件部分*/
// 文件
type Cloud189File struct {
	CreateDate string `json:"createDate"`
	FileCata   int64  `json:"fileCata"`
	Icon       struct {
		//iconOption 5
		SmallUrl string `json:"smallUrl"`
		LargeUrl string `json:"largeUrl"`

		// iconOption 10
		Max600    string `json:"max600"`
		MediumURL string `json:"mediumUrl"`
	} `json:"icon"`
	ID          int64  `json:"id"`
	LastOpTime  string `json:"lastOpTime"`
	Md5         string `json:"md5"`
	MediaType   int    `json:"mediaType"`
	Name        string `json:"name"`
	Orientation int64  `json:"orientation"`
	Rev         string `json:"rev"`
	Size        int64  `json:"size"`
	StarLabel   int64  `json:"starLabel"`
}

// 文件夹
type Cloud189Folder struct {
	ID       int64  `json:"id"`
	ParentID int64  `json:"parentId"`
	Name     string `json:"name"`

	FileCata  int64 `json:"fileCata"`
	FileCount int64 `json:"fileCount"`

	LastOpTime string `json:"lastOpTime"`
	CreateDate string `json:"createDate"`

	FileListSize int64  `json:"fileListSize"`
	Rev          string `json:"rev"`
	StarLabel    int64  `json:"starLabel"`
}

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
	//SrcParentId string `json:"srcParentId"`
}

type CreateUploadFileResult struct {
	// UploadFileId 上传文件请求ID
	UploadFileId int64 `json:"uploadFileId"`
	// FileUploadUrl 上传文件数据的URL路径
	FileUploadUrl string `json:"fileUploadUrl"`
	// FileCommitUrl 上传文件完成后确认路径
	FileCommitUrl string `json:"fileCommitUrl"`
	// FileDataExists 文件是否已存在云盘中，0-未存在，1-已存在
	FileDataExists int `json:"fileDataExists"`
}

type UploadFileStatusResult struct {
	// 上传文件的ID
	UploadFileId int64 `json:"uploadFileId"`
	// 已上传的大小
	DataSize       int64  `json:"dataSize"`
	FileUploadUrl  string `json:"fileUploadUrl"`
	FileCommitUrl  string `json:"fileCommitUrl"`
	FileDataExists int    `json:"fileDataExists"`
}

/*
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
	Code       string          `json:"code"`
	UploadUrls map[string]Part `json:"uploadUrls"`
}

type Part struct {
	RequestURL    string `json:"requestURL"`
	RequestHeader string `json:"requestHeader"`
}
*/
