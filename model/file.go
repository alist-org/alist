package model

import (
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
