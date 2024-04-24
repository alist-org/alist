package webdav

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

func (h *Handler) getModTime(r *http.Request) time.Time {
	return h.getHeaderTime(r, "X-OC-Mtime", "")
}

// owncloud/ nextcloud haven't impl this, but we can add the support since rclone may support this soon.
// try ModTime if CreateTime not found in header
func (h *Handler) getCreateTime(r *http.Request) time.Time {
	return h.getHeaderTime(r, "X-OC-Ctime", "X-OC-Mtime")
}

func (h *Handler) getHeaderTime(r *http.Request, header, alternative string) time.Time {
	hVal := r.Header.Get(header)
	// try alternative
	if hVal == "" && alternative != "" {
		hVal = r.Header.Get(alternative)
	}
	if hVal != "" {
		modTimeUnix, err := strconv.ParseInt(hVal, 10, 64)
		if err == nil {
			return time.Unix(modTimeUnix, 0)
		}
		log.Warnf("getModTime in Webdav, failed to parse %s, %s", header, err)
	}
	return time.Now()
}
