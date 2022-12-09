package search

import (
	"strings"

	"github.com/alist-org/alist/v3/drivers/alist_v3"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/utils"
	mapset "github.com/deckarep/golang-set/v2"
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

func GetIgnorePaths() ([]string, error) {
	storages, err := db.GetEnabledStorages()
	if err != nil {
		return nil, err
	}
	ignorePaths := make([]string, 0)
	var skipDrivers = []string{"AList V2", "AList V3", "Virtual"}
	ignoreAddresses := mapset.NewSet[string]()
	indexAddresses := mapset.NewSet[string]()
	for _, storage := range storages {
		if utils.SliceContains(skipDrivers, storage.Driver) {
			if storage.Driver == "AList V3" {
				addition := alist_v3.Addition{}
				utils.Json.UnmarshalFromString(storage.Addition, &addition)
				if len(addition.Address) > 0 && string(addition.Address[len(addition.Address)-1]) == "/" {
					addition.Address = addition.Address[0 : len(addition.Address)-1]
				}
				if !indexAddresses.Contains(addition.Address) {
					if !ignoreAddresses.Contains(addition.Address) {
						url := addition.Address + "/api/public/settings"
						canIndex := false
						var resp map[string]string
						_, err := base.RestyClient.R().
							SetResult(&resp).Get(url)
						if err == nil {
							if val, ok := resp[conf.AllowIndexed]; ok {
								canIndex = setting.GetBool(val)
							}
						}
						if canIndex {
							indexAddresses.Add(addition.Address)
						} else {
							ignoreAddresses.Add(addition.Address)
						}
					}
					if ignoreAddresses.Contains(addition.Address) {
						ignorePaths = append(ignorePaths, storage.MountPath)
					}
				}
			} else {
				ignorePaths = append(ignorePaths, storage.MountPath)
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
