package aria2

import (
	"github.com/alist-org/alist/v3/pkg/aria2/rpc"
	"github.com/alist-org/alist/v3/pkg/generic_sync"
)

const (
	Downloading = iota
	Paused
	Stopped
	Completed
	Errored
)

type Notify struct {
	Signals generic_sync.MapOf[string, chan int]
}

func NewNotify() *Notify {
	return &Notify{Signals: generic_sync.MapOf[string, chan int]{}}
}

func (n *Notify) OnDownloadStart(events []rpc.Event) {
	for _, e := range events {
		if signal, ok := n.Signals.Load(e.Gid); ok {
			signal <- Downloading
		}
	}
}

func (n *Notify) OnDownloadPause(events []rpc.Event) {
	for _, e := range events {
		if signal, ok := n.Signals.Load(e.Gid); ok {
			signal <- Paused
		}
	}
}

func (n *Notify) OnDownloadStop(events []rpc.Event) {
	for _, e := range events {
		if signal, ok := n.Signals.Load(e.Gid); ok {
			signal <- Stopped
		}
	}
}

func (n *Notify) OnDownloadComplete(events []rpc.Event) {
	for _, e := range events {
		if signal, ok := n.Signals.Load(e.Gid); ok {
			signal <- Completed
		}
	}
}

func (n *Notify) OnDownloadError(events []rpc.Event) {
	for _, e := range events {
		if signal, ok := n.Signals.Load(e.Gid); ok {
			signal <- Errored
		}
	}
}

func (n *Notify) OnBtDownloadComplete(events []rpc.Event) {
	for _, e := range events {
		if signal, ok := n.Signals.Load(e.Gid); ok {
			signal <- Completed
		}
	}
}
