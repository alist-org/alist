package model

import (
	"github.com/Xhofe/alist/conf"
	"sort"
	"strings"
	"time"
)

type File struct {
	Id        string     `json:"-"`
	Name      string     `json:"name"`
	Size      int64      `json:"size"`
	Type      int        `json:"type"`
	Driver    string     `json:"driver"`
	UpdatedAt *time.Time `json:"updated_at"`
	Thumbnail string     `json:"thumbnail"`
	Url       string     `json:"url"`
	SizeStr   string     `json:"size_str"`
	TimeStr   string     `json:"time_str"`
}

func SortFiles(files []File, account *Account) {
	if account.OrderBy == "" {
		return
	}
	sort.Slice(files, func(i, j int) bool {
		switch account.OrderBy {
		case "name":
			{
				c := strings.Compare(files[i].Name, files[j].Name)
				if account.OrderDirection == "DESC" {
					return c >= 0
				}
				return c <= 0
			}
		case "size":
			{
				if account.OrderDirection == "DESC" {
					return files[i].Size >= files[j].Size
				}
				return files[i].Size <= files[j].Size
			}
		case "updated_at":
			if account.OrderDirection == "DESC" {
				return files[i].UpdatedAt.After(*files[j].UpdatedAt)
			}
			return files[i].UpdatedAt.Before(*files[j].UpdatedAt)
		}
		return false
	})
}

func ExtractFolder(files []File, account *Account) {
	if account.ExtractFolder == "" {
		return
	}
	front := account.ExtractFolder == "front"
	sort.SliceStable(files, func(i, j int) bool {
		if files[i].IsDir() || files[j].IsDir() {
			if !files[i].IsDir() {
				return !front
			}
			if !files[j].IsDir() {
				return front
			}
		}
		return false
	})
}

func (f File) GetSize() uint64 {
	return uint64(f.Size)
}

func (f File) GetName() string {
	return f.Name
}

func (f File) ModTime() time.Time {
	return *f.UpdatedAt
}

func (f File) IsDir() bool {
	return f.Type == conf.FOLDER
}

func (f File) GetType() int {
	return f.Type
}
