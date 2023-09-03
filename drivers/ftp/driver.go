package ftp

import (
	"context"
	stdpath "path"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/jlaffaye/ftp"
)

type FTP struct {
	model.Storage
	Addition
	conn *ftp.ServerConn
}

func (d *FTP) Config() driver.Config {
	return config
}

func (d *FTP) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *FTP) Init(ctx context.Context) error {
	return d.login()
}

func (d *FTP) Drop(ctx context.Context) error {
	if d.conn != nil {
		_ = d.conn.Logout()
	}
	return nil
}

func (d *FTP) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	if err := d.login(); err != nil {
		return nil, err
	}
	entries, err := d.conn.List(dir.GetPath())
	if err != nil {
		return nil, err
	}
	res := make([]model.Obj, 0)
	for _, entry := range entries {
		if entry.Name == "." || entry.Name == ".." {
			continue
		}
		f := model.Object{
			Name:     entry.Name,
			Size:     int64(entry.Size),
			Modified: entry.Time,
			IsFolder: entry.Type == ftp.EntryTypeFolder,
		}
		res = append(res, &f)
	}
	return res, nil
}

func (d *FTP) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	if err := d.login(); err != nil {
		return nil, err
	}

	r := NewFileReader(d.conn, file.GetPath(), file.GetSize())
	link := &model.Link{
		MFile: r,
	}
	return link, nil
}

func (d *FTP) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	if err := d.login(); err != nil {
		return err
	}
	return d.conn.MakeDir(stdpath.Join(parentDir.GetPath(), dirName))
}

func (d *FTP) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	if err := d.login(); err != nil {
		return err
	}
	return d.conn.Rename(srcObj.GetPath(), stdpath.Join(dstDir.GetPath(), srcObj.GetName()))
}

func (d *FTP) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	if err := d.login(); err != nil {
		return err
	}
	return d.conn.Rename(srcObj.GetPath(), stdpath.Join(stdpath.Dir(srcObj.GetPath()), newName))
}

func (d *FTP) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotSupport
}

func (d *FTP) Remove(ctx context.Context, obj model.Obj) error {
	if err := d.login(); err != nil {
		return err
	}
	if obj.IsDir() {
		return d.conn.RemoveDirRecur(obj.GetPath())
	} else {
		return d.conn.Delete(obj.GetPath())
	}
}

func (d *FTP) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	if err := d.login(); err != nil {
		return err
	}
	// TODO: support cancel
	return d.conn.Stor(stdpath.Join(dstDir.GetPath(), stream.GetName()), stream)
}

var _ driver.Driver = (*FTP)(nil)
