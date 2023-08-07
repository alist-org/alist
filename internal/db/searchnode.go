package db

import (
	"fmt"
	stdpath "path"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func whereInParent(parent string) *gorm.DB {
	if parent == "/" {
		return db.Where("1 = 1")
	}
	return db.Where(fmt.Sprintf("%s LIKE ?", columnName("parent")),
		fmt.Sprintf("%s/%%", parent)).
		Or(fmt.Sprintf("%s = ?", columnName("parent")), parent)
}

func CreateSearchNode(node *model.SearchNode) error {
	return db.Create(node).Error
}

func BatchCreateSearchNodes(nodes *[]model.SearchNode) error {
	return db.CreateInBatches(nodes, 1000).Error
}

func DeleteSearchNodesByParent(path string) error {
	path = utils.FixAndCleanPath(path)
	err := db.Where(whereInParent(path)).Delete(&model.SearchNode{}).Error
	if err != nil {
		return err
	}
	dir, name := stdpath.Split(path)
	return db.Where(fmt.Sprintf("%s = ? AND %s = ?",
		columnName("parent"), columnName("name")),
		dir, name).Delete(&model.SearchNode{}).Error
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

func SearchNode(req model.SearchReq, useFullText bool) ([]model.SearchNode, int64, error) {
	var searchDB *gorm.DB
	if !useFullText || conf.Conf.Database.Type == "sqlite3" {
		keywordsClause := db.Where("1 = 1")
		for _, keyword := range strings.Fields(req.Keywords) {
			keywordsClause = keywordsClause.Where("name LIKE ?", fmt.Sprintf("%%%s%%", keyword))
		}
		searchDB = db.Model(&model.SearchNode{}).Where(whereInParent(req.Parent)).Where(keywordsClause)
	} else {
		switch conf.Conf.Database.Type {
		case "mysql":
			searchDB = db.Model(&model.SearchNode{}).Where(whereInParent(req.Parent)).
				Where("MATCH (name) AGAINST (? IN BOOLEAN MODE)", "'*"+req.Keywords+"*'")
		case "postgres":
			searchDB = db.Model(&model.SearchNode{}).Where(whereInParent(req.Parent)).
				Where("to_tsvector(name) @@ to_tsquery(?)", strings.Join(strings.Fields(req.Keywords), " & "))
		}
	}

	if req.Scope != 0 {
		isDir := req.Scope == 1
		searchDB.Where(db.Where("is_dir = ?", isDir))
	}

	var count int64
	if err := searchDB.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get search items count")
	}
	var files []model.SearchNode
	if err := searchDB.Order("name asc").Offset((req.Page - 1) * req.PerPage).Limit(req.PerPage).
		Find(&files).Error; err != nil {
		return nil, 0, err
	}
	return files, count, nil
}
