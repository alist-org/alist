package store

import (
	"fmt"
	"github.com/alist-org/alist/v3/conf"
)

func columnName(name string) string {
	if conf.Conf.Database.Type == "postgres" {
		return fmt.Sprintf(`"%s"`, name)
	}
	return fmt.Sprintf("`%s`", name)
}
