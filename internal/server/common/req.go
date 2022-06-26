package common

type PageReq struct {
	PageIndex int `json:"page_index" form:"page_index"`
	PageSize  int `json:"page_size" form:"page_size"`
}
