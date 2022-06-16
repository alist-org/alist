package model

type Meta struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Path      string `json:"path" gorm:"unique" binding:"required"`
	Password  string `json:"password"`
	Hide      string `json:"hide"`
	Upload    bool   `json:"upload"`
	OnlyShows string `json:"only_shows"`
	Readme    string `json:"readme"`
}
