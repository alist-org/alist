package db

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/search/searcher"
)

var config = searcher.Config{
	Name:       "database",
	AutoUpdate: true,
}

func init() {
	searcher.RegisterSearcher(config, func() (searcher.Searcher, error) {
		db := db.GetDb()
		switch conf.Conf.Database.Type {
		case "mysql":
			tableName := fmt.Sprintf("%ssearch_nodes", conf.Conf.Database.TablePrefix)
			tx := db.Exec(fmt.Sprintf("CREATE FULLTEXT INDEX idx_%s_name_fulltext ON %s(name);", tableName, tableName))
			if err := tx.Error; err != nil && !strings.Contains(err.Error(), "Error 1061 (42000)") { // duplicate error
				log.Errorf("failed to create full text index: %v", err)
				return nil, err
			}
		case "postgres":
			db.Exec("CREATE EXTENSION pg_trgm;")
			db.Exec("CREATE EXTENSION btree_gin;")
			tableName := fmt.Sprintf("%ssearch_nodes", conf.Conf.Database.TablePrefix)
			tx := db.Exec(fmt.Sprintf("CREATE INDEX idx_%s_name ON %s USING GIN (name);", tableName, tableName))
			if err := tx.Error; err != nil && !strings.Contains(err.Error(), "SQLSTATE 42P07") {
				log.Errorf("failed to create index using GIN: %v", err)
				return nil, err
			}
		}
		return &DB{}, nil
	})
}
