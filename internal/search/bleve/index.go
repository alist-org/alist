package bleve

import (
	"os"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/blevesearch/bleve/v2"
	log "github.com/sirupsen/logrus"
)

func Init(indexPath *string) bleve.Index {
	fileIndex, err := bleve.Open(*indexPath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Infof("Creating new index...")
		indexMapping := bleve.NewIndexMapping()
		fileIndex, err = bleve.New(*indexPath, indexMapping)
		if err != nil {
			log.Fatal(err)
		}
	}
	return fileIndex
}

func Reset() {
	log.Infof("Removing old index...")
	err := os.RemoveAll(conf.Conf.BleveDir)
	if err != nil {
		log.Fatal(err)
	}
	Init(&conf.Conf.BleveDir)
}
