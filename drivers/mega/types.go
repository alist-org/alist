package mega

import (
	"time"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/t3rm1n4l/go-mega"
)

type MegaNode struct {
	*mega.Node
}

//func (m *MegaNode) GetSize() int64 {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (m *MegaNode) GetName() string {
//	//TODO implement me
//	panic("implement me")
//}

func (m *MegaNode) ModTime() time.Time {
	return m.GetTimeStamp()
}

func (m *MegaNode) IsDir() bool {
	return m.GetType() == mega.FOLDER || m.GetType() == mega.ROOT
}

func (m *MegaNode) GetID() string {
	return m.GetHash()
}

func (m *MegaNode) GetPath() string {
	return ""
}

var _ model.Obj = (*MegaNode)(nil)
