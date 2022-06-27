package data

import (
	"context"
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
}
