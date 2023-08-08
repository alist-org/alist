package crypt

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	stdpath "path"
	"path/filepath"
	"strings"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/net"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/http_range"
)

func RequestRangedHttp(r *http.Request, link *model.Link, offset, length int64) (*http.Response, error) {
	header := net.ProcessHeader(&http.Header{}, &link.Header)
	header = http_range.ApplyRangeToHttpHeader(http_range.Range{Start: offset, Length: length}, header)

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
	}
	return false, true
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

// 139 cloud does not properly return 206 http status code, add a hack here
func checkContentRange(header *http.Header, size, offset int64) bool {
	r, err2 := http_range.ParseRange(header.Get("Content-Range"), size)
	if err2 != nil {
		log.Warnf("Crypt got exception when trying to parse Content-Range, will ignore,err=%s", err2)
	}
	if len(r) == 1 && r[0].Start == offset {
		return true
	}
	return false
}
