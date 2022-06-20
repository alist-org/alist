package aria2

import "github.com/alist-org/alist/v3/pkg/aria2/rpc"

type Notify struct {
}

func (n Notify) OnDownloadStart(events []rpc.Event) {
	//TODO update task status
	panic("implement me")
}

func (n Notify) OnDownloadPause(events []rpc.Event) {
	//TODO update task status
	panic("implement me")
}

func (n Notify) OnDownloadStop(events []rpc.Event) {
	//TODO update task status
	panic("implement me")
}

func (n Notify) OnDownloadComplete(events []rpc.Event) {
	//TODO get files and upload them
	panic("implement me")
}

func (n Notify) OnDownloadError(events []rpc.Event) {
	//TODO update task status
	panic("implement me")
}

func (n Notify) OnBtDownloadComplete(events []rpc.Event) {
	//TODO get files and upload them
	panic("implement me")
}
