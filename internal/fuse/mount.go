package fuse

import "github.com/winfsp/cgofuse/fuse"

func Mount(mountSrc, mountDst string, opts []string) {
	fs := &Fs{RootFolder: mountSrc}
	host := fuse.NewFileSystemHost(fs)
	go host.Mount(mountDst, opts)
}
