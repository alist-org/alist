package model

import (
	"fmt"
	"time"
)

type IndexProgress struct {
	ObjCount     uint64     `json:"obj_count"`
	IsDone       bool       `json:"is_done"`
	LastDoneTime *time.Time `json:"last_done_time"`
	Error        string     `json:"error"`
}

type SearchReq struct {
	Parent   string `json:"parent"`
	Keywords string `json:"keywords"`
	// 0 for all, 1 for dir, 2 for file
	Scope int `json:"scope"`
	PageReq
}

type SearchNode struct {
	Parent string `json:"parent" gorm:"index"`
	Name   string `json:"name"`
	IsDir  bool   `json:"is_dir"`
	Size   int64  `json:"size"`
}

func (p *SearchReq) Validate() error {
	if p.Page < 1 {
		return fmt.Errorf("page can't < 1")
	}
	if p.PerPage < 1 {
		return fmt.Errorf("per_page can't < 1")
	}
	return nil
}

func (s *SearchNode) Type() string {
	return "SearchNode"
}
