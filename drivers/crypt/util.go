package crypt

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/net"
	"github.com/alist-org/alist/v3/internal/op"
	"net/http"
	stdpath "path"
	"path/filepath"
	"strconv"
	"strings"
)

func RequestRangedHttp(r *http.Request, link *model.Link, offset, length int64) (*http.Response, error) {
	header := net.ProcessHeader(&http.Header{}, &link.Header)
	if offset == 0 && length < 0 {
		header.Del("Range")
	} else {
		end := ""
		if length >= 0 {
			end = strconv.FormatInt(offset+length-1, 10)
		}
		header.Set("Range", fmt.Sprintf("bytes=%v-%v", offset, end))
	}

	return net.RequestHttp("GET", header, link.URL)
}

// will give the best guessing based on the path
func guessPath(path string) (isFolder, secondTry bool) {
	if strings.HasSuffix(path, "/") {
		//confirmed a folder
		return true, false
	}
	lastSlash := strings.LastIndex(path, "/")
	if strings.Index(path[lastSlash:], ".") < 0 {
		//no dot, try folder then try file
		return true, true
	} else {
		return false, true
	}
}

func (d *Crypt) getPathForRemote(path string, isFolder bool) (remoteFullPath string) {
	if isFolder && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	dir, fileName := filepath.Split(path)

	remoteDir := d.cipher.EncryptDirName(dir)
	remoteFileName := ""
	if len(strings.TrimSpace(fileName)) > 0 {
		remoteFileName = d.cipher.EncryptFileName(fileName)
	}
	return stdpath.Join(d.RemotePath, remoteDir, remoteFileName)

}

// actual path is used for internal only. any link for user should come from remoteFullPath
func (d *Crypt) getActualPathForRemote(path string, isFolder bool) (string, error) {
	_, remoteActualPath, err := op.GetStorageAndActualPath(d.getPathForRemote(path, isFolder))
	return remoteActualPath, err
}
