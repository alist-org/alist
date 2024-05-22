package lark

import (
	"context"
	"github.com/Xhofe/go-cache"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	log "github.com/sirupsen/logrus"
	"path"
	"time"
)

const objTokenCacheDuration = 5 * time.Minute
const emptyFolderToken = "empty"

var objTokenCache = cache.NewMemCache[string]()
var exOpts = cache.WithEx[string](objTokenCacheDuration)

func (c *Lark) getObjToken(ctx context.Context, folderPath string) (string, bool) {
	if token, ok := objTokenCache.Get(folderPath); ok {
		return token, true
	}

	dir, name := path.Split(folderPath)
	// strip the last slash of dir if it exists
	if len(dir) > 0 && dir[len(dir)-1] == '/' {
		dir = dir[:len(dir)-1]
	}
	if name == "" {
		return c.rootFolderToken, true
	}

	var parentToken string
	var found bool
	parentToken, found = c.getObjToken(ctx, dir)
	if !found {
		return emptyFolderToken, false
	}

	req := larkdrive.NewListFileReqBuilder().FolderToken(parentToken).Build()
	resp, err := c.client.Drive.File.ListByIterator(ctx, req)

	if err != nil {
		log.WithError(err).Error("failed to list files")
		return emptyFolderToken, false
	}

	var file *larkdrive.File
	for {
		found, file, err = resp.Next()
		if !found {
			break
		}

		if err != nil {
			log.WithError(err).Error("failed to get next file")
			break
		}

		if *file.Name == name {
			objTokenCache.Set(folderPath, *file.Token, exOpts)
			return *file.Token, true
		}
	}

	return emptyFolderToken, false
}
