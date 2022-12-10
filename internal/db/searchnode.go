package db

import (
	"fmt"
	"path"
	"strings"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
)

func CreateSearchNode(node *model.SearchNode) error {
	return db.Create(node).Error
}

func BatchCreateSearchNodes(nodes *[]model.SearchNode) error {
	return db.CreateInBatches(nodes, 1000).Error
}

func DeleteSearchNodesByParent(prefix string) error {
	err := db.Where(fmt.Sprintf("%s LIKE ?",
		columnName("parent")), fmt.Sprintf("%s%%", prefix)).
		Delete(&model.SearchNode{}).Error
	if err != nil {
		return err
	}
	dir, name := path.Split(prefix)
	return db.Where(fmt.Sprintf("%s = ? AND %s = ?",
		columnName("parent"), columnName("name")),
		utils.StandardizePath(dir), name).Delete(&model.SearchNode{}).Error
}

func ClearSearchNodes() error {
	return db.Where("1 = 1").Delete(&model.SearchNode{}).Error
}

func GetSearchNodesByParent(parent string) ([]model.SearchNode, error) {
	var nodes []model.SearchNode
	if err := db.Where(fmt.Sprintf("%s = ?",
		columnName("parent")), parent).Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

func SearchNode(req model.SearchReq) ([]model.SearchNode, int64, error) {
	keywordsClause := db.Where("1 = 1")
	for _, keyword := range strings.Split(req.Keywords, " ") {
		keywordsClause = keywordsClause.Where(
			fmt.Sprintf("%s LIKE ?", columnName("name")),
			fmt.Sprintf("%%%s%%", keyword))
	}
	searchDB := db.Model(&model.SearchNode{}).Where(
		fmt.Sprintf("%s LIKE ?", columnName("parent")),
		fmt.Sprintf("%s%%", req.Parent)).Where(keywordsClause)
	var count int64
	if err := searchDB.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get users count")
	}
	var files []model.SearchNode
	if err := searchDB.Offset((req.Page - 1) * req.PerPage).Limit(req.PerPage).Find(&files).Error; err != nil {
		return nil, 0, err
	}
	return files, count, nil
}
