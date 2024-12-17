package tdlib

import (
	"github.com/zelenin/go-tdlib/client"
	"os"
)

func NewListener(tdlibClient *client.Client) *client.Listener {
	listener := tdlibClient.GetListener()

	go func() {
		for update := range listener.Updates {
			switch up := update.(type) {
			case *client.UpdateFile:
				handleFileUpdate(up)
			}
		}
	}()

	return listener
}

func handleFileUpdate(upd *client.UpdateFile) {
	if upd.File.Local == nil {
		return
	}

	path := upd.File.Local.Path
	if path == "" {
		return
	}

	if upd.File.Remote.UploadedSize == upd.File.Size {
		_ = os.Remove(path)
	}
}
