package url_tree

import (
	"context"
	stdpath "path"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type Urls struct {
	model.Storage
	Addition
	root *Node
}

func (d *Urls) Config() driver.Config {
	return config
}

func (d *Urls) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Urls) Init(ctx context.Context) error {
	node, err := BuildTree(d.UrlStructure, d.HeadSize)
	if err != nil {
		return err
	}
	node.calSize()
	d.root = node
	return nil
}

func (d *Urls) Drop(ctx context.Context) error {
	return nil
}

func (d *Urls) Get(ctx context.Context, path string) (model.Obj, error) {
	node := GetNodeFromRootByPath(d.root, path)
	return nodeToObj(node, path)
}

func (d *Urls) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	node := GetNodeFromRootByPath(d.root, dir.GetPath())
	log.Debugf("path: %s, node: %+v", dir.GetPath(), node)
	if node == nil {
		return nil, errs.ObjectNotFound
	}
	if node.isFile() {
		return nil, errs.NotFolder
	}
	return utils.SliceConvert(node.Children, func(node *Node) (model.Obj, error) {
		return nodeToObj(node, stdpath.Join(dir.GetPath(), node.Name))
	})
}

func (d *Urls) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	node := GetNodeFromRootByPath(d.root, file.GetPath())
	log.Debugf("path: %s, node: %+v", file.GetPath(), node)
	if node == nil {
		return nil, errs.ObjectNotFound
	}
	if node.isFile() {
		return &model.Link{
			URL: node.Url,
		}, nil
	}
	return nil, errs.NotFile
}

//func (d *Template) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*Urls)(nil)
