package crypt

import (
	stdpath "path"
	"path/filepath"
	"strings"

	"github.com/alist-org/alist/v3/internal/op"
)

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
