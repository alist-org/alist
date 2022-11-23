package model

import "time"

type Index struct {
	ID              uint      `json:"id" gorm:"primaryKey"`                        // unique key
	MountPath       string    `json:"mount_path" gorm:"unique" binding:"required"` // must be standardized
	Order           int       `json:"order"`                                       // use to sort
	Driver          string    `json:"driver"`                                      // driver used
	CacheExpiration int       `json:"cache_expiration"`                            // cache expire time
	Status          string    `json:"status"`
	Addition        string    `json:"addition" gorm:"type:text"` // Additional information, defined in the corresponding driver
	Remark          string    `json:"remark"`
	Modified        time.Time `json:"modified"`
	Disabled        bool      `json:"disabled"` // if disabled
	Sort
	Proxy
}
