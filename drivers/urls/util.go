package urls

import (
	"fmt"
	stdpath "path"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
)

// build tree from text, text structure:
/**
 * folder1:
 *   name1:url1
 *   url2
 *   folder2:
 *     url3
 *     url4
 *   url5
 * folder3:
 *   url6
 *   url7
 * url8
 */
// if there are no name, use the last segment of url as name
func BuildTree(text string) (*Node, error) {
	lines := strings.Split(text, "\n")
	var root = &Node{Level: -1, Name: "root"}
	stack := []*Node{root}
	for _, line := range lines {
		// calculate indent
		indent := 0
		for i := 0; i < len(line); i++ {
			if line[i] != ' ' {
				break
			}
			indent++
		}
		// if indent is not a multiple of 2, it is an error
		if indent%2 != 0 {
			return nil, fmt.Errorf("the line '%s' is not a multiple of 2", line)
		}
		// calculate level
		level := indent / 2
		line = strings.TrimSpace(line[indent:])
		// if the line is empty, skip
		if line == "" {
			continue
		}
		// if level isn't greater than the level of the top of the stack
		// it is not the child of the top of the stack
		if level <= stack[len(stack)-1].Level {
			// pop the top of the stack
			stack = stack[:len(stack)-1]
		}
		// if the line is a folder
		if isFolder(line) {
			// create a new node
			node := &Node{
				Level: level,
				Name:  strings.TrimSuffix(line, ":"),
			}
			// add the node to the top of the stack
			stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, node)
			// push the node to the stack
			stack = append(stack, node)
		} else {
			// if the line is a file
			// create a new node
			node, err := parseFileLine(line)
			if err != nil {
				return nil, err
			}
			node.Level = level
			// add the node to the top of the stack
			stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, node)
		}
	}
	return root, nil
}

func isFolder(line string) bool {
	return strings.HasSuffix(line, ":")
}

func parseFileLine(line string) (*Node, error) {
	if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
		return &Node{
			Name: stdpath.Base(line),
			Url:  line,
		}, nil
	}
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid line: %s", line)
	}
	return &Node{
		Name: parts[0],
		Url:  parts[1],
	}, nil
}

func splitPath(path string) []string {
	if path == "/" {
		return []string{"root"}
	}
	parts := strings.Split(path, "/")
	parts[0] = "root"
	return parts
}

func GetNodeFromRootByPath(root *Node, path string) *Node {
	return root.getByPath(splitPath(path))
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
