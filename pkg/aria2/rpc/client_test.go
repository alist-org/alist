package rpc

import (
	"context"
	"testing"
	"time"
)

func TestHTTPAll(t *testing.T) {
	const targetURL = "https://nodejs.org/dist/index.json"
	rpc, err := New(context.Background(), "http://localhost:6800/jsonrpc", "", time.Second, &DummyNotifier{})
	if err != nil {
		t.Fatal(err)
	}
	defer rpc.Close()
	g, err := rpc.AddURI([]string{targetURL})
	if err != nil {
		t.Fatal(err)
	}
	println(g)
	if _, err = rpc.TellActive(); err != nil {
		t.Error(err)
	}
	if _, err = rpc.PauseAll(); err != nil {
		t.Error(err)
	}
	if _, err = rpc.TellStatus(g); err != nil {
		t.Error(err)
	}
	if _, err = rpc.GetURIs(g); err != nil {
		t.Error(err)
	}
	if _, err = rpc.GetFiles(g); err != nil {
		t.Error(err)
	}
	if _, err = rpc.GetPeers(g); err != nil {
		t.Error(err)
	}
	if _, err = rpc.TellActive(); err != nil {
		t.Error(err)
	}
	if _, err = rpc.TellWaiting(0, 1); err != nil {
		t.Error(err)
	}
	if _, err = rpc.TellStopped(0, 1); err != nil {
		t.Error(err)
	}
	if _, err = rpc.GetOption(g); err != nil {
		t.Error(err)
	}
	if _, err = rpc.GetGlobalOption(); err != nil {
		t.Error(err)
	}
	if _, err = rpc.GetGlobalStat(); err != nil {
		t.Error(err)
	}
	if _, err = rpc.GetSessionInfo(); err != nil {
		t.Error(err)
	}
	if _, err = rpc.Remove(g); err != nil {
		t.Error(err)
	}
	if _, err = rpc.TellActive(); err != nil {
		t.Error(err)
	}
}

func TestWebsocketAll(t *testing.T) {
	const targetURL = "https://nodejs.org/dist/index.json"
	rpc, err := New(context.Background(), "ws://localhost:6800/jsonrpc", "", time.Second, &DummyNotifier{})
	if err != nil {
		t.Fatal(err)
	}
	defer rpc.Close()
	g, err := rpc.AddURI([]string{targetURL})
	if err != nil {
		t.Fatal(err)
	}
	println(g)
	if _, err = rpc.TellActive(); err != nil {
		t.Error(err)
	}
	if _, err = rpc.PauseAll(); err != nil {
		t.Error(err)
	}
	if _, err = rpc.TellStatus(g); err != nil {
		t.Error(err)
	}
	if _, err = rpc.GetURIs(g); err != nil {
		t.Error(err)
	}
	if _, err = rpc.GetFiles(g); err != nil {
		t.Error(err)
	}
	if _, err = rpc.GetPeers(g); err != nil {
		t.Error(err)
	}
	if _, err = rpc.TellActive(); err != nil {
		t.Error(err)
	}
	if _, err = rpc.TellWaiting(0, 1); err != nil {
		t.Error(err)
	}
	if _, err = rpc.TellStopped(0, 1); err != nil {
		t.Error(err)
	}
	if _, err = rpc.GetOption(g); err != nil {
		t.Error(err)
	}
	if _, err = rpc.GetGlobalOption(); err != nil {
		t.Error(err)
	}
	if _, err = rpc.GetGlobalStat(); err != nil {
		t.Error(err)
	}
	if _, err = rpc.GetSessionInfo(); err != nil {
		t.Error(err)
	}
	if _, err = rpc.Remove(g); err != nil {
		t.Error(err)
	}
	if _, err = rpc.TellActive(); err != nil {
		t.Error(err)
	}
}
