package aria2

import "testing"

func TestConnect(t *testing.T) {
	err := InitAria2Client("http://localhost:16800/jsonrpc", "secret", 3)
	if err != nil {
		t.Errorf("failed to init aria2: %+v", err)
	}
}
