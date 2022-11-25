package bleve

import (
	"os"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/blevesearch/bleve/v2"
	log "github.com/sirupsen/logrus"
)

var index bleve.Index

func Init(indexPath *string) {
	fileIndex, err := bleve.Open(*indexPath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Infof("Creating new index...")
		indexMapping := bleve.NewIndexMapping()
		fileIndex, err = bleve.New(*indexPath, indexMapping)
		if err != nil {
			log.Fatal(err)
		}
	}
	index = fileIndex
	progress := ReadProgress()
	if !progress.IsDone {
		log.Warnf("Last index build does not succeed!")
		WriteProgress(&model.IndexProgress{
			FileCount:    progress.FileCount,
			IsDone:       false,
			LastDoneTime: nil,
		})
	}
}

func Reset() {
	log.Infof("Removing old index...")
	err := os.RemoveAll(conf.Conf.BleveDir)
	if err != nil {
		log.Fatal(err)
	}
	Init(&conf.Conf.BleveDir)
	WriteProgress(&model.IndexProgress{
		FileCount:    0,
		IsDone:       false,
		LastDoneTime: nil,
	})
}
