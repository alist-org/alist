package model

type Meta struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Path      string `json:"path" gorm:"unique" binding:"required"`
	Password  string `json:"password"`
	Upload    bool   `json:"upload"`
	Hide      string `json:"hide"`
	SubFolder bool   `json:"sub_folder"`
	Readme    string `json:"readme"`
}
