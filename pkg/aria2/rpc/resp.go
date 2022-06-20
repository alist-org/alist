//go:generate easyjson -all

package rpc

// StatusInfo represents response of aria2.tellStatus
type StatusInfo struct {
	Gid             string     `json:"gid"`             // GID of the download.
	Status          string     `json:"status"`          // active for currently downloading/seeding downloads. waiting for downloads in the queue; download is not started. paused for paused downloads. error for downloads that were stopped because of error. complete for stopped and completed downloads. removed for the downloads removed by user.
	TotalLength     string     `json:"totalLength"`     // Total length of the download in bytes.
	CompletedLength string     `json:"completedLength"` // Completed length of the download in bytes.
	UploadLength    string     `json:"uploadLength"`    // Uploaded length of the download in bytes.
	BitField        string     `json:"bitfield"`        // Hexadecimal representation of the download progress. The highest bit corresponds to the piece at index 0. Any set bits indicate loaded pieces, while unset bits indicate not yet loaded and/or missing pieces. Any overflow bits at the end are set to zero. When the download was not started yet, this key will not be included in the response.
	DownloadSpeed   string     `json:"downloadSpeed"`   // Download speed of this download measured in bytes/sec.
	UploadSpeed     string     `json:"uploadSpeed"`     // Upload speed of this download measured in bytes/sec.
	InfoHash        string     `json:"infoHash"`        // InfoHash. BitTorrent only.
	NumSeeders      string     `json:"numSeeders"`      // The number of seeders aria2 has connected to. BitTorrent only.
	Seeder          string     `json:"seeder"`          // true if the local endpoint is a seeder. Otherwise, false. BitTorrent only.
	PieceLength     string     `json:"pieceLength"`     // Piece length in bytes.
	NumPieces       string     `json:"numPieces"`       // The number of pieces.
	Connections     string     `json:"connections"`     // The number of peers/servers aria2 has connected to.
	ErrorCode       string     `json:"errorCode"`       // The code of the last error for this item, if any. The value is a string. The error codes are defined in the EXIT STATUS section. This value is only available for stopped/completed downloads.
	ErrorMessage    string     `json:"errorMessage"`    // The (hopefully) human-readable error message associated to errorCode.
	FollowedBy      []string   `json:"followedBy"`      // List of GIDs which are generated as the result of this download. For example, when aria2 downloads a Metalink file, it generates downloads described in the Metalink (see the --follow-metalink option). This value is useful to track auto-generated downloads. If there are no such downloads, this key will not be included in the response.
	BelongsTo       string     `json:"belongsTo"`       // GID of a parent download. Some downloads are a part of another download. For example, if a file in a Metalink has BitTorrent resources, the downloads of ".torrent" files are parts of that parent. If this download has no parent, this key will not be included in the response.
	Dir             string     `json:"dir"`             // Directory to save files.
	Files           []FileInfo `json:"files"`           // Returns the list of files. The elements of this list are the same structs used in aria2.getFiles() method.
	BitTorrent      struct {
		AnnounceList [][]string `json:"announceList"` // List of lists of announce URIs. If the torrent contains announce and no announce-list, announce is converted to the announce-list format.
		Comment      string     `json:"comment"`      // The comment of the torrent. comment.utf-8 is used if available.
		CreationDate int64      `json:"creationDate"` // The creation time of the torrent. The value is an integer since the epoch, measured in seconds.
		Mode         string     `json:"mode"`         // File mode of the torrent. The value is either single or multi.
		Info         struct {
			Name string `json:"name"` // name in info dictionary. name.utf-8 is used if available.
		} `json:"info"` // Struct which contains data from Info dictionary. It contains following keys.
	} `json:"bittorrent"` // Struct which contains information retrieved from the .torrent (file). BitTorrent only. It contains following keys.
}

// URIInfo represents an element of response of aria2.getUris
type URIInfo struct {
	URI    string `json:"uri"`    // URI
	Status string `json:"status"` // 'used' if the URI is in use. 'waiting' if the URI is still waiting in the queue.
}

// FileInfo represents an element of response of aria2.getFiles
type FileInfo struct {
	Index           string    `json:"index"`           // Index of the file, starting at 1, in the same order as files appear in the multi-file torrent.
	Path            string    `json:"path"`            // File path.
	Length          string    `json:"length"`          // File size in bytes.
	CompletedLength string    `json:"completedLength"` // Completed length of this file in bytes. Please note that it is possible that sum of completedLength is less than the completedLength returned by the aria2.tellStatus() method. This is because completedLength in aria2.getFiles() only includes completed pieces. On the other hand, completedLength in aria2.tellStatus() also includes partially completed pieces.
	Selected        string    `json:"selected"`        // true if this file is selected by --select-file option. If --select-file is not specified or this is single-file torrent or not a torrent download at all, this value is always true. Otherwise false.
	URIs            []URIInfo `json:"uris"`            // Returns a list of URIs for this file. The element type is the same struct used in the aria2.getUris() method.
}

// PeerInfo represents an element of response of aria2.getPeers
type PeerInfo struct {
	PeerId        string `json:"peerId"`        // Percent-encoded peer ID.
	IP            string `json:"ip"`            // IP address of the peer.
	Port          string `json:"port"`          // Port number of the peer.
	BitField      string `json:"bitfield"`      // Hexadecimal representation of the download progress of the peer. The highest bit corresponds to the piece at index 0. Set bits indicate the piece is available and unset bits indicate the piece is missing. Any spare bits at the end are set to zero.
	AmChoking     string `json:"amChoking"`     // true if aria2 is choking the peer. Otherwise false.
	PeerChoking   string `json:"peerChoking"`   // true if the peer is choking aria2. Otherwise false.
	DownloadSpeed string `json:"downloadSpeed"` // Download speed (byte/sec) that this client obtains from the peer.
	UploadSpeed   string `json:"uploadSpeed"`   // Upload speed(byte/sec) that this client uploads to the peer.
	Seeder        string `json:"seeder"`        // true if this peer is a seeder. Otherwise false.
}

// ServerInfo represents an element of response of aria2.getServers
type ServerInfo struct {
	Index   string `json:"index"` // Index of the file, starting at 1, in the same order as files appear in the multi-file metalink.
	Servers []struct {
		URI           string `json:"uri"`           // Original URI.
		CurrentURI    string `json:"currentUri"`    // This is the URI currently used for downloading. If redirection is involved, currentUri and uri may differ.
		DownloadSpeed string `json:"downloadSpeed"` // Download speed (byte/sec)
	} `json:"servers"` // A list of structs which contain the following keys.
}

// GlobalStatInfo represents response of aria2.getGlobalStat
type GlobalStatInfo struct {
	DownloadSpeed   string `json:"downloadSpeed"`   // Overall download speed (byte/sec).
	UploadSpeed     string `json:"uploadSpeed"`     // Overall upload speed(byte/sec).
	NumActive       string `json:"numActive"`       // The number of active downloads.
	NumWaiting      string `json:"numWaiting"`      // The number of waiting downloads.
	NumStopped      string `json:"numStopped"`      // The number of stopped downloads in the current session. This value is capped by the --max-download-result option.
	NumStoppedTotal string `json:"numStoppedTotal"` // The number of stopped downloads in the current session and not capped by the --max-download-result option.
}

// VersionInfo represents response of aria2.getVersion
type VersionInfo struct {
	Version  string   `json:"version"`         // Version number of aria2 as a string.
	Features []string `json:"enabledFeatures"` // List of enabled features. Each feature is given as a string.
}

// SessionInfo represents response of aria2.getSessionInfo
type SessionInfo struct {
	Id string `json:"sessionId"` // Session ID, which is generated each time when aria2 is invoked.
}

// Method is an element of parameters used in system.multicall
type Method struct {
	Name   string        `json:"methodName"` // Method name to call
	Params []interface{} `json:"params"`     // Array containing parameters to the method call
}
