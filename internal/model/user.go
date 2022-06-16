package model

const (
	GENERAL = iota
	GUEST   // only one exists
	ADMIN
)

type User struct {
	ID          uint   `json:"id" gorm:"primaryKey"` // unique key
	Name        string `json:"name" gorm:"unique"`   // username
	Password    string `json:"password"`             // password
	BasePath    string `json:"base_path"`            // base path
	AllowUpload bool   `json:"allow_upload"`         // allow upload
	Role        int    `json:"role"`                 // user's role
}

func (u User) IsGuest() bool {
	return u.Role == GUEST
}

func (u User) IsAdmin() bool {
	return u.Role == ADMIN
}
