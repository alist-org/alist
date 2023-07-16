package model

import (
	"time"

	"github.com/alist-org/alist/v3/pkg/utils"
)

type ObjWrapName struct {
	Name string
	Obj
}

func (o *ObjWrapName) Unwrap() Obj {
	return o.Obj
}

func (o *ObjWrapName) GetName() string {
	if o.Name == "" {
		o.Name = utils.MappingName(o.Obj.GetName())
	}
	return o.Name
}

type Object struct {
	ID       string
	Path     string
	Name     string
	Size     int64
	Modified time.Time
	IsFolder bool
	Hash     string
	HashType string
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

func (o *Object) SetPath(path string) {
	o.Path = path
}

func (o *Object) SetHash(hash string, hashType string) {
	o.Hash = hash
	o.HashType = hashType
}

func (o *Object) GetHash() (string, string) {
	return o.Hash, o.HashType
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
