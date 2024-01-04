package template

type ListResp struct {
	Msg       string     `json:"msg"`
	Total     int        `json:"total"`
	Code      int        `json:"code"`
	Offset    int        `json:"offset"`
	TotalPage int        `json:"totalPage"`
	Limit     int        `json:"limit"`
	List      []ListItem `json:"list"`
}

type ListItem struct {
	IconId         int         `json:"iconId"`
	IsAmt          int         `json:"isAmt"`
	FolderDesc     string      `json:"folderDesc,omitempty"`
	AddTime        string      `json:"addTime"`
	FolderId       int64       `json:"folderId"`
	ParentId       int64       `json:"parentId"`
	ParentName     string      `json:"parentName"`
	NoteType       int         `json:"noteType,omitempty"`
	UpdTime        string      `json:"updTime"`
	IsShare        int         `json:"isShare"`
	FolderIcon     string      `json:"folderIcon,omitempty"`
	FolderName     string      `json:"folderName,omitempty"`
	FileType       int         `json:"fileType"`
	Status         int         `json:"status"`
	IsFileShare    int         `json:"isFileShare,omitempty"`
	FileName       string      `json:"fileName,omitempty"`
	FileStars      float64     `json:"fileStars,omitempty"`
	IsFileDownload int         `json:"isFileDownload,omitempty"`
	FileComments   int         `json:"fileComments,omitempty"`
	FileSize       int64       `json:"fileSize,omitempty"`
	FileIcon       string      `json:"fileIcon,omitempty"`
	FileDownloads  int         `json:"fileDownloads,omitempty"`
	FileUrl        interface{} `json:"fileUrl"`
	FileLikes      int         `json:"fileLikes,omitempty"`
	FileId         int64       `json:"fileId,omitempty"`
}

type Part struct {
	PartNumber int    `json:"partNumber"`
	ETag       string `json:"etag"`
}

type UploadResultResp struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	List []struct {
		FileIconId int    `json:"fileIconId"`
		FileName   string `json:"fileName"`
		FileIcon   string `json:"fileIcon"`
		FileId     int64  `json:"fileId"`
		Status     int    `json:"status"`
		Token      string `json:"token"`
	} `json:"list"`
}
