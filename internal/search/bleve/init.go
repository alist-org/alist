package bleve

import (
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/search/searcher"
	"github.com/blevesearch/bleve/v2"
	log "github.com/sirupsen/logrus"
)

var config = searcher.Config{
	Name: "bleve",
}

func Init(indexPath *string) (bleve.Index, error) {
	log.Debugf("bleve path: %s", *indexPath)
	fileIndex, err := bleve.Open(*indexPath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Infof("Creating new index...")
		indexMapping := bleve.NewIndexMapping()
		searchNodeMapping := bleve.NewDocumentMapping()
		searchNodeMapping.AddFieldMappingsAt("is_dir", bleve.NewBooleanFieldMapping())
		// TODO: appoint analyzer
		parentFieldMapping := bleve.NewTextFieldMapping()
		searchNodeMapping.AddFieldMappingsAt("parent", parentFieldMapping)
		// TODO: appoint analyzer
		nameFieldMapping := bleve.NewKeywordFieldMapping()
		searchNodeMapping.AddFieldMappingsAt("name", nameFieldMapping)
		indexMapping.AddDocumentMapping("SearchNode", searchNodeMapping)
		fileIndex, err = bleve.New(*indexPath, indexMapping)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return fileIndex, nil
}

func init() {
	searcher.RegisterSearcher(config, func() (searcher.Searcher, error) {
		b, err := Init(&conf.Conf.BleveDir)
		if err != nil {
			return nil, err
		}
		return &Bleve{BIndex: b}, nil
	})
}
