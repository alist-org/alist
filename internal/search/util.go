package search

import (
	"strings"

	"github.com/alist-org/alist/v3/drivers/alist_v3"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func Progress() (*model.IndexProgress, error) {
	p := setting.GetStr(conf.IndexProgress)
	var progress model.IndexProgress
	err := utils.Json.UnmarshalFromString(p, &progress)
	return &progress, err
}

func WriteProgress(progress *model.IndexProgress) {
	p, err := utils.Json.MarshalToString(progress)
	if err != nil {
		log.Errorf("marshal progress error: %+v", err)
	}
	err = db.SaveSettingItem(model.SettingItem{
		Key:   conf.IndexProgress,
		Value: p,
		Type:  conf.TypeText,
		Group: model.SINGLE,
		Flag:  model.PRIVATE,
	})
	if err != nil {
		log.Errorf("save progress error: %+v", err)
	}
}

func GetIndexPaths() []string {
	indexPaths := make([]string, 0)
	customIndexPaths := setting.GetStr(conf.IndexPaths)
	if customIndexPaths != "" {
		indexPaths = append(indexPaths, strings.Split(customIndexPaths, "\n")...)
	}
	return indexPaths
}

func isIndexPath(path string, indexPaths []string) bool {
	for _, indexPaths := range indexPaths {
		if strings.HasPrefix(path, indexPaths) {
			return true
		}
	}
	return false
}

func GetIgnorePaths() ([]string, error) {
	storages := op.GetAllStorages()
	ignorePaths := make([]string, 0)
	var skipDrivers = []string{"AList V2", "AList V3", "Virtual"}
	v3Visited := make(map[string]bool)
	for _, storage := range storages {
		if utils.SliceContains(skipDrivers, storage.Config().Name) {
			if storage.Config().Name == "AList V3" {
				addition := storage.GetAddition().(*alist_v3.Addition)
				allowIndexed, visited := v3Visited[addition.Address]
				if !visited {
					url := addition.Address + "/api/public/settings"
					res, err := base.RestyClient.R().Get(url)
					if err == nil {
						allowIndexed = utils.Json.Get(res.Body(), "data", conf.AllowIndexed).ToBool()
					}
				}
				if allowIndexed {
					ignorePaths = append(ignorePaths, storage.GetStorage().MountPath)
				}
			} else {
				ignorePaths = append(ignorePaths, storage.GetStorage().MountPath)
			}
		}
	}
	customIgnorePaths := setting.GetStr(conf.IgnorePaths)
	if customIgnorePaths != "" {
		ignorePaths = append(ignorePaths, strings.Split(customIgnorePaths, "\n")...)
	}
	return ignorePaths, nil
}

func isIgnorePath(path string, ignorePaths []string) bool {
	for _, ignorePath := range ignorePaths {
		if strings.HasPrefix(path, ignorePath) {
			return true
		}
	}
	return false
}
