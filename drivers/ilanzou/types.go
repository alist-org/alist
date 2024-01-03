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
	IconId     int    `json:"iconId"`
	IsAmt      int    `json:"isAmt"`
	FolderDesc string `json:"folderDesc"`
	AddTime    string `json:"addTime"`
	FolderId   int    `json:"folderId"`
	ParentId   int    `json:"parentId"`
	NoteType   int    `json:"noteType"`
	UpdTime    string `json:"updTime"`
	IsShare    int    `json:"isShare"`
	FolderIcon string `json:"folderIcon"`
	FolderName string `json:"folderName"`
	FileType   int    `json:"fileType"`
	Status     int    `json:"status"`
}
