package data

import (
	"context"
	"github.com/alist-org/alist/v3/cmd/args"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/message"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	log "github.com/sirupsen/logrus"
)

func initDevData() {
	err := operations.CreateAccount(context.Background(), model.Account{
		VirtualPath: "/",
		Index:       0,
		Driver:      "Local",
		Status:      "",
		Addition:    `{"root_folder":"."}`,
	})
	if err != nil {
		log.Fatalf("failed to create account: %+v", err)
	}
	err = db.CreateUser(&model.User{
		Username:   "Noah",
		Password:   "hsu",
		BasePath:   "/data",
		Role:       0,
		Permission: 512,
	})
	if err != nil {
		log.Fatalf("failed to create user: %+v", err)
	}
}

func initDevDo() {
	if args.Dev {
		go func() {
			err := message.GetMessenger().WaitSend(map[string]string{
				"type": "dev",
				"msg":  "dev mode",
			}, 10)
			if err != nil {
				log.Debugf("%+v", err)
			}
			m, err := message.GetMessenger().WaitReceive(10)
			if err != nil {
				log.Debugf("%+v", err)
			} else {
				log.Debugf("received: %+v", m)
			}
		}()
	}
}
