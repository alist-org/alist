package model

import "time"

type File struct {
	Name      string     `json:"name"`
	Size      int64      `json:"size"`
	Type      int        `json:"type"`
	UpdatedAt *time.Time `json:"updated_at"`
}
