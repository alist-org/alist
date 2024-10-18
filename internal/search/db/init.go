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
		var parser string
		switch conf.Conf.Database.Type {
		case "mysql":
			if conf.Conf.Database.Parser != "" {
				parser = fmt.Sprintf(" WITH PARSER %s", conf.Conf.Database.Parser)
			}
			tableName := fmt.Sprintf("%ssearch_nodes", conf.Conf.Database.TablePrefix)
			tx := db.Exec(fmt.Sprintf("CREATE FULLTEXT INDEX idx_%s_name_fulltext%s ON %s(name);", tableName, parser, tableName))
			if err := tx.Error; err != nil && !strings.Contains(err.Error(), "Error 1061 (42000)") { // duplicate error
				log.Errorf("failed to create full text index: %v", err)
				return nil, err
			}
		case "postgres":
			if conf.Conf.Database.Parser != "" {
				parser = fmt.Sprintf(" gin_%s", conf.Conf.Database.Parser)
			}
			db.Exec("CREATE EXTENSION pg_trgm;")
			db.Exec("CREATE EXTENSION btree_gin;")
			tableName := fmt.Sprintf("%ssearch_nodes", conf.Conf.Database.TablePrefix)
			tx := db.Exec(fmt.Sprintf("CREATE INDEX idx_%s_name ON %s USING GIN (name%s);", tableName, tableName, parser))
			if err := tx.Error; err != nil && !strings.Contains(err.Error(), "SQLSTATE 42P07") {
				log.Errorf("failed to create index using GIN: %v", err)
				return nil, err
			}
		}
		return &DB{}, nil
	})
}
