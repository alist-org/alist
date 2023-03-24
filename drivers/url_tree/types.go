package url_tree

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
