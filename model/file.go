package model

import "time"

type File struct {
	Name      string     `json:"name"`
	Size      int64      `json:"size"`
	Type      int        `json:"type"`
	Driver    string     `json:"driver"`
	UpdatedAt *time.Time `json:"updated_at"`
	Thumbnail string     `json:"thumbnail"`
}
