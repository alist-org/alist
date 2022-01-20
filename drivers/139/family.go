package _39

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	jsoniter "github.com/json-iterator/go"
	"path"
)

func (driver Cloud139) familyGetFiles(catalogID string, account *model.Account) ([]model.File, error) {
	pageNum := 1
	files := make([]model.File, 0)
	for {
		data := newJson(base.Json{
			"catalogID":       catalogID,
			"contentSortType": 0,
			"pageInfo": base.Json{
				"pageNum":  pageNum,
				"pageSize": 100,
			},
			"sortDirection": 1,
		}, account)

		var resp QueryContentListResp
		_, err := driver.Post("/orchestration/familyCloud/content/v1.0/queryContentList", data, &resp, account)
		if err != nil {
			return nil, err
		}
		for _, catalog := range resp.Data.CloudCatalogList {
			f := model.File{
				Id:        catalog.CatalogID,
				Name:      catalog.CatalogName,
				Size:      0,
				Type:      conf.FOLDER,
				Driver:    driver.Config().Name,
				UpdatedAt: getTime(catalog.LastUpdateTime),
			}
			files = append(files, f)
		}
		for _, content := range resp.Data.CloudContentList {
			f := model.File{
				Id:        content.ContentID,
				Name:      content.ContentName,
				Size:      content.ContentSize,
				Type:      utils.GetFileType(path.Ext(content.ContentName)),
				Driver:    driver.Config().Name,
				UpdatedAt: getTime(content.LastUpdateTime),
				Thumbnail: content.ThumbnailURL,
				//Thumbnail: content.BigthumbnailURL,
			}
			files = append(files, f)
		}
		if 100*pageNum > resp.Data.TotalCount {
			break
		}
		pageNum++
	}
	return files, nil
}

func (driver Cloud139) familyLink(contentId string, account *model.Account) (string, error) {
	data := newJson(base.Json{
		"contentID": contentId,
		//"path":"",
	}, account)
	res, err := driver.Post("/orchestration/familyCloud/content/v1.0/getFileDownLoadURL",
		data, nil, account)
	if err != nil {
		return "", err
	}
	return jsoniter.Get(res, "data", "downloadURL").ToString(), nil
}
