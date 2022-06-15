package rpc

import (
	log "github.com/sirupsen/logrus"
)

type Event struct {
	Gid string `json:"gid"` // GID of the download
}

// The RPC server might send notifications to the client.
// Notifications is unidirectional, therefore the client which receives the notification must not respond to it.
// The method signature of a notification is much like a normal method request but lacks the id key

type websocketResponse struct {
	clientResponse
	Method string  `json:"method"`
	Params []Event `json:"params"`
}

// Notifier handles rpc notification from aria2 server
type Notifier interface {
	// OnDownloadStart will be sent when a download is started.
	OnDownloadStart([]Event)
	// OnDownloadPause will be sent when a download is paused.
	OnDownloadPause([]Event)
	// OnDownloadStop will be sent when a download is stopped by the user.
	OnDownloadStop([]Event)
	// OnDownloadComplete will be sent when a download is complete. For BitTorrent downloads, this notification is sent when the download is complete and seeding is over.
	OnDownloadComplete([]Event)
	// OnDownloadError will be sent when a download is stopped due to an error.
	OnDownloadError([]Event)
	// OnBtDownloadComplete will be sent when a torrent download is complete but seeding is still going on.
	OnBtDownloadComplete([]Event)
}

type DummyNotifier struct{}

func (DummyNotifier) OnDownloadStart(events []Event)      { log.Printf("%s started.", events) }
func (DummyNotifier) OnDownloadPause(events []Event)      { log.Printf("%s paused.", events) }
func (DummyNotifier) OnDownloadStop(events []Event)       { log.Printf("%s stopped.", events) }
func (DummyNotifier) OnDownloadComplete(events []Event)   { log.Printf("%s completed.", events) }
func (DummyNotifier) OnDownloadError(events []Event)      { log.Printf("%s error.", events) }
func (DummyNotifier) OnBtDownloadComplete(events []Event) { log.Printf("bt %s completed.", events) }
