package _123Link

import (
	"fmt"
	url2 "net/url"
	stdpath "path"
	"strconv"
	"strings"
	"time"
)

// build tree from text, text structure definition:
/**
 * FolderName:
 *   [FileSize:][Modified:]Url
 */
/**
 * For example:
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
		for level <= stack[len(stack)-1].Level {
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

// line definition:
// [FileSize:][Modified:]Url
func parseFileLine(line string) (*Node, error) {
	// if there is no url, it is an error
	if !strings.Contains(line, "http://") && !strings.Contains(line, "https://") {
		return nil, fmt.Errorf("invalid line: %s, because url is required for file", line)
	}
	index := strings.Index(line, "http://")
	if index == -1 {
		index = strings.Index(line, "https://")
	}
	url := line[index:]
	info := line[:index]
	node := &Node{
		Url: url,
	}
	name := stdpath.Base(url)
	unescape, err := url2.PathUnescape(name)
	if err == nil {
		name = unescape
	}
	node.Name = name
	if index > 0 {
		if !strings.HasSuffix(info, ":") {
			return nil, fmt.Errorf("invalid line: %s, because file info must end with ':'", line)
		}
		info = info[:len(info)-1]
		if info == "" {
			return nil, fmt.Errorf("invalid line: %s, because file name can't be empty", line)
		}
		infoParts := strings.Split(info, ":")
		size, err := strconv.ParseInt(infoParts[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid line: %s, because file size must be an integer", line)
		}
		node.Size = size
		if len(infoParts) > 1 {
			modified, err := strconv.ParseInt(infoParts[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid line: %s, because file modified must be an unix timestamp", line)
			}
			node.Modified = modified
		} else {
			node.Modified = time.Now().Unix()
		}
	}
	return node, nil
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
