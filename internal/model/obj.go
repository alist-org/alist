package model

import (
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/maruel/natural"
)

type ObjUnwrap interface {
	Unwrap() Obj
}

type Obj interface {
	GetSize() int64
	GetName() string
	ModTime() time.Time
	IsDir() bool
	//GetHash() (string, string)

	// The internal information of the driver.
	// If you want to use it, please understand what it means
	GetID() string
	GetPath() string
}

type FileStreamer interface {
	io.ReadCloser
	Obj
	GetMimetype() string
	SetReadCloser(io.ReadCloser)
	NeedStore() bool
	GetReadCloser() io.ReadCloser
	GetOld() Obj
}

type URL interface {
	URL() string
}

type Thumb interface {
	Thumb() string
}

type SetPath interface {
	SetPath(path string)
}
type SetHash interface {
	SetHash(hash string, hashType string)
}

func SortFiles(objs []Obj, orderBy, orderDirection string) {
	if orderBy == "" {
		return
	}
	sort.Slice(objs, func(i, j int) bool {
		switch orderBy {
		case "name":
			{
				c := natural.Less(objs[i].GetName(), objs[j].GetName())
				if orderDirection == "desc" {
					return !c
				}
				return c
			}
		case "size":
			{
				if orderDirection == "desc" {
					return objs[i].GetSize() >= objs[j].GetSize()
				}
				return objs[i].GetSize() <= objs[j].GetSize()
			}
		case "modified":
			if orderDirection == "desc" {
				return objs[i].ModTime().After(objs[j].ModTime())
			}
			return objs[i].ModTime().Before(objs[j].ModTime())
		}
		return false
	})
}

func ExtractFolder(objs []Obj, extractFolder string) {
	if extractFolder == "" {
		return
	}
	front := extractFolder == "front"
	sort.SliceStable(objs, func(i, j int) bool {
		if objs[i].IsDir() || objs[j].IsDir() {
			if !objs[i].IsDir() {
				return !front
			}
			if !objs[j].IsDir() {
				return front
			}
		}
		return false
	})
}

func WrapObjName(objs Obj) Obj {
	return &ObjWrapName{Obj: objs}
}

func WrapObjsName(objs []Obj) {
	for i := 0; i < len(objs); i++ {
		objs[i] = &ObjWrapName{Obj: objs[i]}
	}
}

func UnwrapObj(obj Obj) Obj {
	if unwrap, ok := obj.(ObjUnwrap); ok {
		obj = unwrap.Unwrap()
	}
	return obj
}

func GetThumb(obj Obj) (thumb string, ok bool) {
	if obj, ok := obj.(Thumb); ok {
		return obj.Thumb(), true
	}
	if unwrap, ok := obj.(ObjUnwrap); ok {
		return GetThumb(unwrap.Unwrap())
	}
	return thumb, false
}

func GetUrl(obj Obj) (url string, ok bool) {
	if obj, ok := obj.(URL); ok {
		return obj.URL(), true
	}
	if unwrap, ok := obj.(ObjUnwrap); ok {
		return GetUrl(unwrap.Unwrap())
	}
	return url, false
}

// Merge
func NewObjMerge() *ObjMerge {
	return &ObjMerge{
		set: mapset.NewSet[string](),
	}
}

type ObjMerge struct {
	regs []*regexp.Regexp
	set  mapset.Set[string]
}

func (om *ObjMerge) Merge(objs []Obj, objs_ ...Obj) []Obj {
	newObjs := make([]Obj, 0, len(objs)+len(objs_))
	newObjs = om.insertObjs(om.insertObjs(newObjs, objs...), objs_...)
	return newObjs
}

func (om *ObjMerge) insertObjs(objs []Obj, objs_ ...Obj) []Obj {
	for _, obj := range objs_ {
		if om.clickObj(obj) {
			objs = append(objs, obj)
		}
	}
	return objs
}

func (om *ObjMerge) clickObj(obj Obj) bool {
	for _, reg := range om.regs {
		if reg.MatchString(obj.GetName()) {
			return false
		}
	}
	return om.set.Add(obj.GetName())
}

func (om *ObjMerge) InitHideReg(hides string) {
	rs := strings.Split(hides, "\n")
	om.regs = make([]*regexp.Regexp, 0, len(rs))
	for _, r := range rs {
		om.regs = append(om.regs, regexp.MustCompile(r))
	}
}

func (om *ObjMerge) Reset() {
	om.set.Clear()
}
