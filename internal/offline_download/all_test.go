package offline_download

import "testing"

func TestGetFiles(t *testing.T) {
	files, err := GetFiles("..")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		t.Log(file.Name, file.Size, file.Path)
	}
}
