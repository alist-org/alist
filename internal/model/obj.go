package model

import (
	"io"
	"sort"
	"strings"
	"time"
)

type Obj interface {
	GetSize() int64
	GetName() string
	ModTime() time.Time
	IsDir() bool
	GetID() string
}

type FileStreamer interface {
	io.ReadCloser
	Obj
	GetMimetype() string
	SetReadCloser(io.ReadCloser)
	NeedStore() bool
	GetReadCloser() io.ReadCloser
}

type URL interface {
	URL() string
}

type Thumbnail interface {
	Thumbnail() string
}

type SetID interface {
	SetID(id string)
}

func SortFiles(objs []Obj, orderBy, orderDirection string) {
	if orderBy == "" {
		return
	}
	sort.Slice(objs, func(i, j int) bool {
		switch orderBy {
		case "name":
			{
				c := strings.Compare(objs[i].GetName(), objs[j].GetName())
				if orderDirection == "desc" {
					return c >= 0
				}
				return c <= 0
			}
		case "size":
			{
				if orderDirection == "desc" {
					return objs[i].GetSize() >= objs[j].GetSize()
				}
				return objs[i].GetSize() <= objs[j].GetSize()
			}
		case "updated_at":
			if orderDirection == "desc" {
				return objs[i].ModTime().After(objs[j].ModTime())
			}
			return objs[i].ModTime().Before(objs[j].ModTime())
		}
		return false
	})
}

func ExtractFolder(objs []Obj, extractFolder string) {
	if extractFolder == "" {
		return
	}
	front := extractFolder == "front"
	sort.SliceStable(objs, func(i, j int) bool {
		if objs[i].IsDir() || objs[j].IsDir() {
			if !objs[i].IsDir() {
				return !front
			}
			if !objs[j].IsDir() {
				return front
			}
		}
		return false
	})
}
