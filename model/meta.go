package model

type Meta struct {
	Path     string `json:"path" gorm:"primaryKey"`
	Password string `json:"password"`
	Hide     bool   `json:"hide"`
	Ignore   bool   `json:"ignore"`
}
