package ipfs

import (
	"context"
	"fmt"
	"net/url"
	stdpath "path"
	"path/filepath"
	"strings"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	shell "github.com/ipfs/go-ipfs-api"
)

type IPFS struct {
	model.Storage
	Addition
	sh      *shell.Shell
	gateURL *url.URL
}

func (d *IPFS) Config() driver.Config {
	return config
}

func (d *IPFS) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *IPFS) Init(ctx context.Context) error {
	d.sh = shell.NewShell(d.Endpoint)
	gateURL, err := url.Parse(d.Gateway)
	if err != nil {
		return err
	}
	d.gateURL = gateURL
	return nil
}

func (d *IPFS) Drop(ctx context.Context) error {
	return nil
}

func (d *IPFS) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	path := dir.GetPath()
	if path[len(path):] != "/" {
		path += "/"
	}

	path_cid, err := d.sh.FilesStat(ctx, path)
	if err != nil {
		return nil, err
	}

	dirs, err := d.sh.List(path_cid.Hash)
	if err != nil {
		return nil, err
	}

	objlist := []model.Obj{}
	for _, file := range dirs {
		gateurl := *d.gateURL
		gateurl.Path = "ipfs/" + file.Hash
		gateurl.RawQuery = "filename=" + url.PathEscape(file.Name)
		objlist = append(objlist, &model.ObjectURL{
			Object: model.Object{ID: file.Hash, Name: file.Name, Size: int64(file.Size), IsFolder: file.Type == 1},
			Url:    model.Url{Url: gateurl.String()},
		})
	}

	return objlist, nil
}

func (d *IPFS) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	link := d.Gateway + "/ipfs/" + file.GetID() + "/?filename=" + url.PathEscape(file.GetName())
	return &model.Link{URL: link}, nil
}

func (d *IPFS) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	path := parentDir.GetPath()
	if path[len(path):] != "/" {
		path += "/"
	}
	return d.sh.FilesMkdir(ctx, path+dirName)
}

func (d *IPFS) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return d.sh.FilesMv(ctx, srcObj.GetPath(), dstDir.GetPath())
}

func (d *IPFS) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	newFileName := filepath.Dir(srcObj.GetPath()) + "/" + newName
	return d.sh.FilesMv(ctx, srcObj.GetPath(), strings.ReplaceAll(newFileName, "\\", "/"))
}

func (d *IPFS) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj, optional
	fmt.Println(srcObj.GetPath())
	fmt.Println(dstDir.GetPath())
	newFileName := dstDir.GetPath() + "/" + filepath.Base(srcObj.GetPath())
	fmt.Println(newFileName)
	return d.sh.FilesCp(ctx, srcObj.GetPath(), strings.ReplaceAll(newFileName, "\\", "/"))
}

func (d *IPFS) Remove(ctx context.Context, obj model.Obj) error {
	// TODO remove obj, optional
	return d.sh.FilesRm(ctx, obj.GetPath(), true)
}

func (d *IPFS) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// TODO upload file, optional
	_, err := d.sh.Add(stream, ToFiles(stdpath.Join(dstDir.GetPath(), stream.GetName())))
	return err
}

func ToFiles(dstDir string) shell.AddOpts {
	return func(rb *shell.RequestBuilder) error {
		rb.Option("to-files", dstDir)
		return nil
	}
}

//func (d *Template) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*IPFS)(nil)
