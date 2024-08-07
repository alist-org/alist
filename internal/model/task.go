package model

type TaskItem struct {
	Key         string `json:"key"`
	PersistData string `gorm:"type:text" json:"persist_data"`
}
