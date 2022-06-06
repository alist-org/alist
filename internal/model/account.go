package model

type Account struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	VirtualPath string `json:"virtual_path"`
	Index       int    `json:"index"`
	Driver      string `json:"driver"`
	Status      string `json:"status"`
	Custom      string `json:"custom"`
}
