package model

type Account struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	VirtualPath string `json:"virtual_path"`
	Index       int    `json:"index"`
	Driver      string `json:"driver"`
	Status      string `json:"status"`
	Addition    string `json:"addition"`
	Remark      string `json:"remark"`
	Sort
	Proxy
}

type Sort struct {
	OrderBy        string `json:"order_by"`
	OrderDirection string `json:"order_direction"`
	ExtractFolder  string `json:"extract_folder"`
}

type Proxy struct {
	WebProxy     string `json:"web_proxy"`
	WebdavProxy  bool   `json:"webdav_proxy"`
	WebdavDirect bool   `json:"webdav_direct"`
	DownProxyUrl string `json:"down_proxy_url"`
}
