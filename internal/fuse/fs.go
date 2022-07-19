package fuse

import "github.com/winfsp/cgofuse/fuse"

type Fs struct {
	RootFolder string
	fuse.FileSystemBase
}

func (fs *Fs) Init() {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Destroy() {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Statfs(path string, stat *fuse.Statfs_t) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Mknod(path string, mode uint32, dev uint64) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Mkdir(path string, mode uint32) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Unlink(path string) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Rmdir(path string) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Link(oldpath string, newpath string) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Symlink(target string, newpath string) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Readlink(path string) (int, string) {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Rename(oldpath string, newpath string) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Chmod(path string, mode uint32) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Chown(path string, uid uint32, gid uint32) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Utimens(path string, tmsp []fuse.Timespec) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Access(path string, mask uint32) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Create(path string, flags int, mode uint32) (int, uint64) {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Open(path string, flags int) (int, uint64) {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Getattr(path string, stat *fuse.Stat_t, fh uint64) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Truncate(path string, size int64, fh uint64) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Read(path string, buff []byte, ofst int64, fh uint64) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Write(path string, buff []byte, ofst int64, fh uint64) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Flush(path string, fh uint64) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Release(path string, fh uint64) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Fsync(path string, datasync bool, fh uint64) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Opendir(path string) (int, uint64) {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Readdir(path string, fill func(name string, stat *fuse.Stat_t, ofst int64) bool, ofst int64, fh uint64) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Releasedir(path string, fh uint64) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Fsyncdir(path string, datasync bool, fh uint64) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Setxattr(path string, name string, value []byte, flags int) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Getxattr(path string, name string) (int, []byte) {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Removexattr(path string, name string) int {
	//TODO implement me
	panic("implement me")
}

func (fs *Fs) Listxattr(path string, fill func(name string) bool) int {
	//TODO implement me
	panic("implement me")
}

var _ fuse.FileSystemInterface = (*Fs)(nil)
