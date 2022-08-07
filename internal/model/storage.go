package model

import "time"

type Storage struct {
	ID              uint      `json:"id" gorm:"primaryKey"`                        // unique key
	MountPath       string    `json:"mount_path" gorm:"unique" binding:"required"` // must be standardized
	Index           int       `json:"index"`                                       // use to sort
	Driver          string    `json:"driver"`                                      // driver used
	CacheExpiration int       `json:"cache_expiration"`                            // cache expire time
	Status          string    `json:"status"`
	Addition        string    `json:"addition" gorm:"type:text"` // Additional information, defined in the corresponding driver
	Remark          string    `json:"remark"`
	Modified        time.Time `json:"modified"`
	Sort
	Proxy
}

type Sort struct {
	OrderBy        string `json:"order_by"`
	OrderDirection string `json:"order_direction"`
	ExtractFolder  string `json:"extract_folder"`
}

type Proxy struct {
	WebProxy     bool   `json:"web_proxy"`
	WebdavPolicy string `json:"webdav_policy"`
	DownProxyUrl string `json:"down_proxy_url"`
}

func (a *Storage) GetStorage() Storage {
	return *a
}

func (a *Storage) SetStatus(status string) {
	a.Status = status
}

func (p Proxy) Webdav302() bool {
	return p.WebdavPolicy == "302_redirect"
}

func (p Proxy) WebdavProxy() bool {
	return p.WebdavPolicy == "use_proxy_url"
}

func (p Proxy) WebdavNative() bool {
	return !p.Webdav302() && !p.WebdavProxy()
}
