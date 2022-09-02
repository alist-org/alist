package model

import "time"

type Object struct {
	ID       string
	Path     string
	Name     string
	Size     int64
	Modified time.Time
	IsFolder bool
}

func (o *Object) GetName() string {
	return o.Name
}

func (o *Object) GetSize() int64 {
	return o.Size
}

func (o *Object) ModTime() time.Time {
	return o.Modified
}

func (o *Object) IsDir() bool {
	return o.IsFolder
}

func (o *Object) GetID() string {
	return o.ID
}

func (o *Object) GetPath() string {
	return o.Path
}

func (o *Object) SetPath(id string) {
	o.Path = id
}

type Thumbnail struct {
	Thumbnail string
}

type Url struct {
	Url string
}

func (w Url) URL() string {
	return w.Url
}

func (t Thumbnail) Thumb() string {
	return t.Thumbnail
}

type ObjThumb struct {
	Object
	Thumbnail
}

type ObjectURL struct {
	Object
	Url
}

type ObjThumbURL struct {
	Object
	Thumbnail
	Url
}
