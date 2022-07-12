package common

type PageReq struct {
	PageIndex int `json:"page_index" form:"page_index"`
	PageSize  int `json:"page_size" form:"page_size"`
}

const MaxUint = ^uint(0)
const MinUint = 0
const MaxInt = int(MaxUint >> 1)
const MinInt = -MaxInt - 1

func (p *PageReq) Validate() {
	if p.PageIndex < 1 {
		p.PageIndex = 1
	}
	if p.PageSize < 1 {
		p.PageSize = MaxInt
	}
}
