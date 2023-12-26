package _123Link

import (
	"time"

	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
)

// Node is a node in the folder tree
type Node struct {
	Url      string
	Name     string
	Level    int
	Modified int64
	Size     int64
	Children []*Node
}

func (node *Node) getByPath(paths []string) *Node {
	if len(paths) == 0 || node == nil {
		return nil
	}
	if node.Name != paths[0] {
		return nil
	}
	if len(paths) == 1 {
		return node
	}
	for _, child := range node.Children {
		tmp := child.getByPath(paths[1:])
		if tmp != nil {
			return tmp
		}
	}
	return nil
}

func (node *Node) isFile() bool {
	return node.Url != ""
}

func (node *Node) calSize() int64 {
	if node.isFile() {
		return node.Size
	}
	var size int64 = 0
	for _, child := range node.Children {
		size += child.calSize()
	}
	node.Size = size
	return size
}

func nodeToObj(node *Node, path string) (model.Obj, error) {
	if node == nil {
		return nil, errs.ObjectNotFound
	}
	return &model.Object{
		Name:     node.Name,
		Size:     node.Size,
		Modified: time.Unix(node.Modified, 0),
		IsFolder: !node.isFile(),
		Path:     path,
	}, nil
}
