package model

import "time"

type Object struct {
	ID       string
	Name     string
	Size     uint64
	Modified time.Time
	IsFolder bool
}

func (f Object) GetName() string {
	return f.Name
}

func (f Object) GetSize() uint64 {
	return f.Size
}

func (f Object) ModTime() time.Time {
	return f.Modified
}

func (f Object) IsDir() bool {
	return f.IsFolder
}

func (f Object) GetID() string {
	return f.ID
}
