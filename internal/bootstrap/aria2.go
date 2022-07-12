package bootstrap

import (
	"github.com/alist-org/alist/v3/internal/aria2"
	log "github.com/sirupsen/logrus"
)

func InitAria2() {
	go func() {
		_, err := aria2.InitClient(2)
		log.Errorf("failed to init aria2 client: %+v", err)
	}()
}
