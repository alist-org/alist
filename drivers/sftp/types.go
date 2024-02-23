package sftp

import (
	"os"
	stdpath "path"
	"strings"

	"github.com/alist-org/alist/v3/internal/model"
	log "github.com/sirupsen/logrus"
)

func (d *SFTP) fileToObj(f os.FileInfo, dir string) (model.Obj, error) {
	symlink := f.Mode()&os.ModeSymlink != 0
	if !symlink {
		return &model.Object{
			Name:     f.Name(),
			Size:     f.Size(),
			Modified: f.ModTime(),
			IsFolder: f.IsDir(),
		}, nil
	}
	path := stdpath.Join(dir, f.Name())
	// set target path
	target, err := d.client.ReadLink(path)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(target, "/") {
		target = stdpath.Join(dir, target)
	}
	_f, err := d.client.Stat(target)
	if err != nil {
		if d.IgnoreSymlinkError {
			return &model.Object{
				Name:     f.Name(),
				Size:     f.Size(),
				Modified: f.ModTime(),
				IsFolder: f.IsDir(),
			}, nil
		}
		return nil, err
	}
	// set basic info
	obj := &model.Object{
		Name:     f.Name(),
		Size:     _f.Size(),
		Modified: _f.ModTime(),
		IsFolder: _f.IsDir(),
		Path:     target,
	}
	log.Debugf("[sftp] obj: %+v, is symlink: %v", obj, symlink)
	return obj, nil
}
