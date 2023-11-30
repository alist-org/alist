package model

import "time"

type DirCache struct {
	Path     string    `json:"path" gorm:"index"`
	Modified time.Time `json:"modified"`
	Size     int64     `json:"size"`
}
