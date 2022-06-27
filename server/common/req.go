package common

type PageReq struct {
	PageIndex int `json:"page_index" form:"page_index"`
	PageSize  int `json:"page_size" form:"page_size"`
}

func (p *PageReq) Validate() {
	if p.PageIndex < 1 {
		p.PageIndex = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 50
	}
}
