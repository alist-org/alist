package yaegi_storage

import "time"

type ObjectHelper struct {
	IValue   any
	ID       string
	Path     string
	Name     string
	Size     int64
	Modified time.Time
	IsFolder bool
}

func (o *ObjectHelper) GetName() string {
	return o.Name
}

func (o *ObjectHelper) GetSize() int64 {
	return o.Size
}

func (o *ObjectHelper) ModTime() time.Time {
	return o.Modified
}

func (o *ObjectHelper) IsDir() bool {
	return o.IsFolder
}

func (o *ObjectHelper) GetID() string {
	return o.ID
}

func (o *ObjectHelper) GetPath() string {
	return o.Path
}

func (o *ObjectHelper) SetPath(id string) {
	o.Path = id
}
