package model

type Account struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	VirtualPath string `json:"virtual_path"`
	Index       int    `json:"index"`
	Type        string `json:"type"`
	Status      string `json:"status"`
}
