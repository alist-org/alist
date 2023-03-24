package url_tree_test

import (
	"testing"

	"github.com/alist-org/alist/v3/drivers/url_tree"
)

func testTree() (*url_tree.Node, error) {
	text := `folder1:
  name1:url1
  url2
  folder2:
    url3
    url4
  url5
folder3:
  url6
  url7
url8`
	return url_tree.BuildTree(text, false)
}

func TestBuildTree(t *testing.T) {
	node, err := testTree()
	if err != nil {
		t.Errorf("failed to build tree: %+v", err)
	} else {
		t.Logf("tree: %+v", node)
	}
}

func TestGetNode(t *testing.T) {
	root, err := testTree()
	if err != nil {
		t.Errorf("failed to build tree: %+v", err)
		return
	}
	node := url_tree.GetNodeFromRootByPath(root, "/")
	if node != root {
		t.Errorf("got wrong node: %+v", node)
	}
	url3 := url_tree.GetNodeFromRootByPath(root, "/folder1/folder2/url3")
	if url3 != root.Children[0].Children[2].Children[0] {
		t.Errorf("got wrong node: %+v", url3)
	}
}
