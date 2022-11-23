package index

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type Progress struct {
	FileCount    uint64     `json:"file_count"`
	IsDone       bool       `json:"is_done"`
	LastDoneTime *time.Time `json:"last_done_time"`
}

func ReadProgress() Progress {
	progressFilePath := filepath.Join(conf.Conf.IndexDir, "progress.json")
	_, err := os.Stat(progressFilePath)
	progress := Progress{0, false, nil}
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

func WriteProgress(progress *Progress) {
	progressFilePath := filepath.Join(conf.Conf.IndexDir, "progress.json")
	log.Infof("write index progress: %v", progress)
	if !utils.WriteJsonToFile(progressFilePath, progress) {
		log.Fatalf("failed to write to index progress file")
	}
}
