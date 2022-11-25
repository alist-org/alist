package bleve

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func ReadProgress() model.IndexProgress {
	progressFilePath := filepath.Join(conf.Conf.BleveDir, "progress.json")
	_, err := os.Stat(progressFilePath)
	progress := model.IndexProgress{}
	if errors.Is(err, os.ErrNotExist) {
		if !utils.WriteJsonToFile(progressFilePath, progress) {
			log.Fatalf("failed to create index progress file")
		}
	}
	progressBytes, err := os.ReadFile(progressFilePath)
	if err != nil {
		log.Fatalf("reading index progress file error: %+v", err)
	}
	err = utils.Json.Unmarshal(progressBytes, &progress)
	if err != nil {
		log.Fatalf("load index progress error: %+v", err)
	}
	return progress
}

func WriteProgress(progress *model.IndexProgress) {
	progressFilePath := filepath.Join(conf.Conf.BleveDir, "progress.json")
	log.Infof("write index progress: %v", progress)
	if !utils.WriteJsonToFile(progressFilePath, progress) {
		log.Fatalf("failed to write to index progress file")
	}
}
