package db

import (
	"fmt"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

func CreateSearchNode(node *model.SearchNode) error {
	return db.Create(node).Error
}

func DeleteSearchNodesByParent(parent string) error {
	return db.Where(fmt.Sprintf("%s LIKE ?",
		columnName("path")), fmt.Sprintf("%s%%", parent)).
		Delete(&model.SearchNode{}).Error
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
	searchDB := db.Model(&model.SearchNode{}).Where(
		fmt.Sprintf("%s LIKE ? AND %s LIKE ?",
			columnName("parent"),
			columnName("name")),
		fmt.Sprintf("%s%%", req.Parent),
		fmt.Sprintf("%%%s%%", req.Keywords))
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
