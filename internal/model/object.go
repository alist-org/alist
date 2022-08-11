package model

import "time"

type Object struct {
	ID       string
	Name     string
	Size     int64
	Modified time.Time
	IsFolder bool
}

func (o Object) GetName() string {
	return o.Name
}

func (o Object) GetSize() int64 {
	return o.Size
}

func (o Object) ModTime() time.Time {
	return o.Modified
}

func (o Object) IsDir() bool {
	return o.IsFolder
}

func (o Object) GetID() string {
	return o.ID
}

func (o *Object) SetID(id string) {
	o.ID = id
}

type Thumbnail struct {
	Thumbnail string
}

func (t Thumbnail) Thumb() string {
	return t.Thumbnail
}

type ObjectThumbnail struct {
	Object
	Thumbnail
}
