package urls_test

import (
	"testing"

	"github.com/alist-org/alist/v3/drivers/urls"
)

func testTree() (*urls.Node, error) {
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
	return urls.BuildTree(text)
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
	node := urls.GetNodeFromRootByPath(root, "/")
	if node != root {
		t.Errorf("got wrong node: %+v", node)
	}
	url3 := urls.GetNodeFromRootByPath(root, "/folder1/folder2/url3")
	if url3 != root.Children[0].Children[2].Children[0] {
		t.Errorf("got wrong node: %+v", url3)
	}
}
