package model

import "time"

type File struct {
	ID       string
	Name     string
	Size     uint64
	Modified time.Time
	IsFolder bool
}

func (f File) GetName() string {
	return f.Name
}

func (f File) GetSize() uint64 {
	return f.Size
}

func (f File) ModTime() time.Time {
	return f.Modified
}

func (f File) IsDir() bool {
	return f.IsFolder
}

func (f File) GetID() string {
	return f.ID
}
