package template

import (
	"time"

	"github.com/xhofe/wopan-sdk-go"
)

// do others that not defined in Driver interface

func (d *Wopan) getSortRule() int {
	switch d.SortRule {
	case "name_asc":
		return wopan.SortNameAsc
	case "name_desc":
		return wopan.SortNameDesc
	case "time_asc":
		return wopan.SortTimeAsc
	case "time_desc":
		return wopan.SortTimeDesc
	case "size_asc":
		return wopan.SortSizeAsc
	case "size_desc":
		return wopan.SortSizeDesc
	default:
		return wopan.SortNameAsc
	}
}

func (d *Wopan) getSpaceType() string {
	if d.FamilyID == "" {
		return wopan.SpaceTypePersonal
	}
	return wopan.SpaceTypeFamily
}

// 20230607214351
func getTime(str string) (time.Time, error) {
	return time.Parse("20060102150405", str)
}
