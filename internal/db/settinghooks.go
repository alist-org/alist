package db

import (
	"regexp"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type SettingItemHook struct {
	Hook func(item *model.SettingItem) error
}

var SettingItemHooks = map[string]SettingItemHook{
	conf.VideoTypes: {
		Hook: func(item *model.SettingItem) error {
			conf.TypesMap[conf.VideoTypes] = strings.Split(item.Value, ",")
			return nil
		},
	},
	conf.AudioTypes: {
		Hook: func(item *model.SettingItem) error {
			conf.TypesMap[conf.AudioTypes] = strings.Split(item.Value, ",")
			return nil
		},
	},
	conf.ImageTypes: {
		Hook: func(item *model.SettingItem) error {
			conf.TypesMap[conf.ImageTypes] = strings.Split(item.Value, ",")
			return nil
		},
	},
	conf.TextTypes: {
		Hook: func(item *model.SettingItem) error {
			conf.TypesMap[conf.TextTypes] = strings.Split(item.Value, ",")
			return nil
		},
	},
	//conf.OfficeTypes: {
	//	Hook: func(item *model.SettingItem) error {
	//		conf.TypesMap[conf.OfficeTypes] = strings.Split(item.Value, ",")
	//		return nil
	//	},
	//},
	conf.ProxyTypes: {
		func(item *model.SettingItem) error {
			conf.TypesMap[conf.ProxyTypes] = strings.Split(item.Value, ",")
			return nil
		},
	},
	conf.PrivacyRegs: {
		Hook: func(item *model.SettingItem) error {
			regStrs := strings.Split(item.Value, "\n")
			regs := make([]*regexp.Regexp, 0, len(regStrs))
			for _, regStr := range regStrs {
				reg, err := regexp.Compile(regStr)
				if err != nil {
					return errors.WithStack(err)
				}
				regs = append(regs, reg)
			}
			conf.PrivacyReg = regs
			return nil
		},
	},
	conf.FilenameCharMapping: {
		Hook: func(item *model.SettingItem) error {
			err := utils.Json.UnmarshalFromString(item.Value, &conf.FilenameCharMap)
			if err != nil {
				return err
			}
			log.Debugf("filename char mapping: %+v", conf.FilenameCharMap)
			return nil
		},
	},
}

func HandleSettingItem(item *model.SettingItem) (bool, error) {
	if hook, ok := SettingItemHooks[item.Key]; ok {
		return true, hook.Hook(item)
	}
	return false, nil
}

// func HandleSettingItems(items []model.SettingItem) error {
// 	for i := range items {
// 		if err := HandleSettingItem(&items[i]); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
