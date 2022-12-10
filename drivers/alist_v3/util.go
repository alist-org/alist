package alist_v3

import (
	"errors"

	"github.com/alist-org/alist/v3/server/common"
)

func checkResp(resp common.Resp[interface{}], err error) error {
	if err != nil {
		return err
	}
	if resp.Message == "success" {
		return nil
	}
	return errors.New(resp.Message)
}
