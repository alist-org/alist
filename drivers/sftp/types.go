package template

import "time"

// write all struct here

type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type File struct {
	Id        string     `json:"id"`
	FileName  string     `json:"file_name"`
	Size      int64      `json:"size"`
	File      bool       `json:"file"`
	UpdatedAt *time.Time `json:"updated_at"`
}
