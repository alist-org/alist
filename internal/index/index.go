package index

import (
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
		WriteProgress(&Progress{
			FileCount:    progress.FileCount,
			IsDone:       false,
			LastDoneTime: nil,
		})
	}
}
