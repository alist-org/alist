package rpc

import (
	"context"
	"encoding/base64"
	"errors"
	"net/url"
	"os"
	"time"
)

// Option is a container for specifying Call parameters and returning results
type Option map[string]interface{}

type Client interface {
	Protocol
	Close() error
}

type client struct {
	caller
	url   *url.URL
	token string
}

var (
	errInvalidParameter = errors.New("invalid parameter")
	errNotImplemented   = errors.New("not implemented")
	errConnTimeout      = errors.New("connect to aria2 daemon timeout")
)

// New returns an instance of Client
func New(ctx context.Context, uri string, token string, timeout time.Duration, notifier Notifier) (Client, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	var caller caller
	switch u.Scheme {
	case "http", "https":
		caller = newHTTPCaller(ctx, u, timeout, notifier)
	case "ws", "wss":
		caller, err = newWebsocketCaller(ctx, u.String(), timeout, notifier)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errInvalidParameter
	}
	c := &client{caller: caller, url: u, token: token}
	return c, nil
}

// `aria2.addUri([secret, ]uris[, options[, position]])`
// This method adds a new download. uris is an array of HTTP/FTP/SFTP/BitTorrent URIs (strings) pointing to the same resource.
// If you mix URIs pointing to different resources, then the download may fail or be corrupted without aria2 complaining.
// When adding BitTorrent Magnet URIs, uris must have only one element and it should be BitTorrent Magnet URI.
// options is a struct and its members are pairs of option name and value.
// If position is given, it must be an integer starting from 0.
// The new download will be inserted at position in the waiting queue.
// If position is omitted or position is larger than the current size of the queue, the new download is appended to the end of the queue.
// This method returns the GID of the newly registered download.
func (c *client) AddURI(uris []string, options ...interface{}) (gid string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, uris)
	if options != nil {
		params = append(params, options...)
	}
	err = c.Call(aria2AddURI, params, &gid)
	return
}

// `aria2.addTorrent([secret, ]torrent[, uris[, options[, position]]])`
// This method adds a BitTorrent download by uploading a ".torrent" file.
// If you want to add a BitTorrent Magnet URI, use the aria2.addUri() method instead.
// torrent must be a base64-encoded string containing the contents of the ".torrent" file.
// uris is an array of URIs (string). uris is used for Web-seeding.
// For single file torrents, the URI can be a complete URI pointing to the resource; if URI ends with /, name in torrent file is added.
// For multi-file torrents, name and path in torrent are added to form a URI for each file. options is a struct and its members are pairs of option name and value.
// If position is given, it must be an integer starting from 0.
// The new download will be inserted at position in the waiting queue.
// If position is omitted or position is larger than the current size of the queue, the new download is appended to the end of the queue.
// This method returns the GID of the newly registered download.
// If --rpc-save-upload-metadata is true, the uploaded data is saved as a file named as the hex string of SHA-1 hash of data plus ".torrent" in the directory specified by --dir option.
// E.g. a file name might be 0a3893293e27ac0490424c06de4d09242215f0a6.torrent.
// If a file with the same name already exists, it is overwritten!
// If the file cannot be saved successfully or --rpc-save-upload-metadata is false, the downloads added by this method are not saved by --save-session.
func (c *client) AddTorrent(filename string, options ...interface{}) (gid string, err error) {
	co, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	file := base64.StdEncoding.EncodeToString(co)
	params := make([]interface{}, 0, 3)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, file)
	params = append(params, []interface{}{})
	if options != nil {
		params = append(params, options...)
	}
	err = c.Call(aria2AddTorrent, params, &gid)
	return
}

// `aria2.addMetalink([secret, ]metalink[, options[, position]])`
// This method adds a Metalink download by uploading a ".metalink" file.
// metalink is a base64-encoded string which contains the contents of the ".metalink" file.
// options is a struct and its members are pairs of option name and value.
// If position is given, it must be an integer starting from 0.
// The new download will be inserted at position in the waiting queue.
// If position is omitted or position is larger than the current size of the queue, the new download is appended to the end of the queue.
// This method returns an array of GIDs of newly registered downloads.
// If --rpc-save-upload-metadata is true, the uploaded data is saved as a file named hex string of SHA-1 hash of data plus ".metalink" in the directory specified by --dir option.
// E.g. a file name might be 0a3893293e27ac0490424c06de4d09242215f0a6.metalink.
// If a file with the same name already exists, it is overwritten!
// If the file cannot be saved successfully or --rpc-save-upload-metadata is false, the downloads added by this method are not saved by --save-session.
func (c *client) AddMetalink(filename string, options ...interface{}) (gid []string, err error) {
	co, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	file := base64.StdEncoding.EncodeToString(co)
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, file)
	if options != nil {
		params = append(params, options...)
	}
	err = c.Call(aria2AddMetalink, params, &gid)
	return
}

// `aria2.remove([secret, ]gid)`
// This method removes the download denoted by gid (string).
// If the specified download is in progress, it is first stopped.
// The status of the removed download becomes removed.
// This method returns GID of removed download.
func (c *client) Remove(gid string) (g string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2Remove, params, &g)
	return
}

// `aria2.forceRemove([secret, ]gid)`
// This method removes the download denoted by gid.
// This method behaves just like aria2.remove() except that this method removes the download without performing any actions which take time, such as contacting BitTorrent trackers to unregister the download first.
func (c *client) ForceRemove(gid string) (g string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2ForceRemove, params, &g)
	return
}

// `aria2.pause([secret, ]gid)`
// This method pauses the download denoted by gid (string).
// The status of paused download becomes paused.
// If the download was active, the download is placed in the front of waiting queue.
// While the status is paused, the download is not started.
// To change status to waiting, use the aria2.unpause() method.
// This method returns GID of paused download.
func (c *client) Pause(gid string) (g string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2Pause, params, &g)
	return
}

// `aria2.pauseAll([secret])`
// This method is equal to calling aria2.pause() for every active/waiting download.
// This methods returns OK.
func (c *client) PauseAll() (ok string, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2PauseAll, params, &ok)
	return
}

// `aria2.forcePause([secret, ]gid)`
// This method pauses the download denoted by gid.
// This method behaves just like aria2.pause() except that this method pauses downloads without performing any actions which take time, such as contacting BitTorrent trackers to unregister the download first.
func (c *client) ForcePause(gid string) (g string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2ForcePause, params, &g)
	return
}

// `aria2.forcePauseAll([secret])`
// This method is equal to calling aria2.forcePause() for every active/waiting download.
// This methods returns OK.
func (c *client) ForcePauseAll() (ok string, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2ForcePauseAll, params, &ok)
	return
}

// `aria2.unpause([secret, ]gid)`
// This method changes the status of the download denoted by gid (string) from paused to waiting, making the download eligible to be restarted.
// This method returns the GID of the unpaused download.
func (c *client) Unpause(gid string) (g string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2Unpause, params, &g)
	return
}

// `aria2.unpauseAll([secret])`
// This method is equal to calling aria2.unpause() for every active/waiting download.
// This methods returns OK.
func (c *client) UnpauseAll() (ok string, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2UnpauseAll, params, &ok)
	return
}

// `aria2.tellStatus([secret, ]gid[, keys])`
// This method returns the progress of the download denoted by gid (string).
// keys is an array of strings.
// If specified, the response contains only keys in the keys array.
// If keys is empty or omitted, the response contains all keys.
// This is useful when you just want specific keys and avoid unnecessary transfers.
// For example, aria2.tellStatus("2089b05ecca3d829", ["gid", "status"]) returns the gid and status keys only.
// The response is a struct and contains following keys. Values are strings.
// https://aria2.github.io/manual/en/html/aria2c.html#aria2.tellStatus
func (c *client) TellStatus(gid string, keys ...string) (info StatusInfo, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	if keys != nil {
		params = append(params, keys)
	}
	err = c.Call(aria2TellStatus, params, &info)
	return
}

// `aria2.getUris([secret, ]gid)`
// This method returns the URIs used in the download denoted by gid (string).
// The response is an array of structs and it contains following keys. Values are string.
//
//	uri        URI
//	status    'used' if the URI is in use. 'waiting' if the URI is still waiting in the queue.
func (c *client) GetURIs(gid string) (infos []URIInfo, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2GetURIs, params, &infos)
	return
}

// `aria2.getFiles([secret, ]gid)`
// This method returns the file list of the download denoted by gid (string).
// The response is an array of structs which contain following keys. Values are strings.
// https://aria2.github.io/manual/en/html/aria2c.html#aria2.getFiles
func (c *client) GetFiles(gid string) (infos []FileInfo, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2GetFiles, params, &infos)
	return
}

// `aria2.getPeers([secret, ]gid)`
// This method returns a list peers of the download denoted by gid (string).
// This method is for BitTorrent only.
// The response is an array of structs and contains the following keys. Values are strings.
// https://aria2.github.io/manual/en/html/aria2c.html#aria2.getPeers
func (c *client) GetPeers(gid string) (infos []PeerInfo, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2GetPeers, params, &infos)
	return
}

// `aria2.getServers([secret, ]gid)`
// This method returns currently connected HTTP(S)/FTP/SFTP servers of the download denoted by gid (string).
// The response is an array of structs and contains the following keys. Values are strings.
// https://aria2.github.io/manual/en/html/aria2c.html#aria2.getServers
func (c *client) GetServers(gid string) (infos []ServerInfo, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2GetServers, params, &infos)
	return
}

// `aria2.tellActive([secret][, keys])`
// This method returns a list of active downloads.
// The response is an array of the same structs as returned by the aria2.tellStatus() method.
// For the keys parameter, please refer to the aria2.tellStatus() method.
func (c *client) TellActive(keys ...string) (infos []StatusInfo, err error) {
	params := make([]interface{}, 0, 1)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	if keys != nil {
		params = append(params, keys)
	}
	err = c.Call(aria2TellActive, params, &infos)
	return
}

// `aria2.tellWaiting([secret, ]offset, num[, keys])`
// This method returns a list of waiting downloads, including paused ones.
// offset is an integer and specifies the offset from the download waiting at the front.
// num is an integer and specifies the max. number of downloads to be returned.
// For the keys parameter, please refer to the aria2.tellStatus() method.
// If offset is a positive integer, this method returns downloads in the range of [offset, offset + num).
// offset can be a negative integer. offset == -1 points last download in the waiting queue and offset == -2 points the download before the last download, and so on.
// Downloads in the response are in reversed order then.
// For example, imagine three downloads "A","B" and "C" are waiting in this order.
// aria2.tellWaiting(0, 1) returns ["A"].
// aria2.tellWaiting(1, 2) returns ["B", "C"].
// aria2.tellWaiting(-1, 2) returns ["C", "B"].
// The response is an array of the same structs as returned by aria2.tellStatus() method.
func (c *client) TellWaiting(offset, num int, keys ...string) (infos []StatusInfo, err error) {
	params := make([]interface{}, 0, 3)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, offset)
	params = append(params, num)
	if keys != nil {
		params = append(params, keys)
	}
	err = c.Call(aria2TellWaiting, params, &infos)
	return
}

// `aria2.tellStopped([secret, ]offset, num[, keys])`
// This method returns a list of stopped downloads.
// offset is an integer and specifies the offset from the least recently stopped download.
// num is an integer and specifies the max. number of downloads to be returned.
// For the keys parameter, please refer to the aria2.tellStatus() method.
// offset and num have the same semantics as described in the aria2.tellWaiting() method.
// The response is an array of the same structs as returned by the aria2.tellStatus() method.
func (c *client) TellStopped(offset, num int, keys ...string) (infos []StatusInfo, err error) {
	params := make([]interface{}, 0, 3)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, offset)
	params = append(params, num)
	if keys != nil {
		params = append(params, keys)
	}
	err = c.Call(aria2TellStopped, params, &infos)
	return
}

// `aria2.changePosition([secret, ]gid, pos, how)`
// This method changes the position of the download denoted by gid in the queue.
// pos is an integer. how is a string.
// If how is POS_SET, it moves the download to a position relative to the beginning of the queue.
// If how is POS_CUR, it moves the download to a position relative to the current position.
// If how is POS_END, it moves the download to a position relative to the end of the queue.
// If the destination position is less than 0 or beyond the end of the queue, it moves the download to the beginning or the end of the queue respectively.
// The response is an integer denoting the resulting position.
// For example, if GID#2089b05ecca3d829 is currently in position 3, aria2.changePosition('2089b05ecca3d829', -1, 'POS_CUR') will change its position to 2. Additionally aria2.changePosition('2089b05ecca3d829', 0, 'POS_SET') will change its position to 0 (the beginning of the queue).
func (c *client) ChangePosition(gid string, pos int, how string) (p int, err error) {
	params := make([]interface{}, 0, 3)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	params = append(params, pos)
	params = append(params, how)
	err = c.Call(aria2ChangePosition, params, &p)
	return
}

// `aria2.changeUri([secret, ]gid, fileIndex, delUris, addUris[, position])`
// This method removes the URIs in delUris from and appends the URIs in addUris to download denoted by gid.
// delUris and addUris are lists of strings.
// A download can contain multiple files and URIs are attached to each file.
// fileIndex is used to select which file to remove/attach given URIs. fileIndex is 1-based.
// position is used to specify where URIs are inserted in the existing waiting URI list. position is 0-based.
// When position is omitted, URIs are appended to the back of the list.
// This method first executes the removal and then the addition.
// position is the position after URIs are removed, not the position when this method is called.
// When removing an URI, if the same URIs exist in download, only one of them is removed for each URI in delUris.
// In other words, if there are three URIs http://example.org/aria2 and you want remove them all, you have to specify (at least) 3 http://example.org/aria2 in delUris.
// This method returns a list which contains two integers.
// The first integer is the number of URIs deleted.
// The second integer is the number of URIs added.
func (c *client) ChangeURI(gid string, fileindex int, delUris []string, addUris []string, position ...int) (p []int, err error) {
	params := make([]interface{}, 0, 5)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	params = append(params, fileindex)
	params = append(params, delUris)
	params = append(params, addUris)
	if position != nil {
		params = append(params, position[0])
	}
	err = c.Call(aria2ChangeURI, params, &p)
	return
}

// `aria2.getOption([secret, ]gid)`
// This method returns options of the download denoted by gid.
// The response is a struct where keys are the names of options.
// The values are strings.
// Note that this method does not return options which have no default value and have not been set on the command-line, in configuration files or RPC methods.
func (c *client) GetOption(gid string) (m Option, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2GetOption, params, &m)
	return
}

// `aria2.changeOption([secret, ]gid, options)`
// This method changes options of the download denoted by gid (string) dynamically. options is a struct.
// The following options are available for active downloads:
//
//	bt-max-peers
//	bt-request-peer-speed-limit
//	bt-remove-unselected-file
//	force-save
//	max-download-limit
//	max-upload-limit
//
// For waiting or paused downloads, in addition to the above options, options listed in Input File subsection are available, except for following options: dry-run, metalink-base-uri, parameterized-uri, pause, piece-length and rpc-save-upload-metadata option.
// This method returns OK for success.
func (c *client) ChangeOption(gid string, option Option) (ok string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	if option != nil {
		params = append(params, option)
	}
	err = c.Call(aria2ChangeOption, params, &ok)
	return
}

// `aria2.getGlobalOption([secret])`
// This method returns the global options.
// The response is a struct.
// Its keys are the names of options.
// Values are strings.
// Note that this method does not return options which have no default value and have not been set on the command-line, in configuration files or RPC methods. Because global options are used as a template for the options of newly added downloads, the response contains keys returned by the aria2.getOption() method.
func (c *client) GetGlobalOption() (m Option, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2GetGlobalOption, params, &m)
	return
}

// `aria2.changeGlobalOption([secret, ]options)`
// This method changes global options dynamically.
// options is a struct.
// The following options are available:
//
//	bt-max-open-files
//	download-result
//	log
//	log-level
//	max-concurrent-downloads
//	max-download-result
//	max-overall-download-limit
//	max-overall-upload-limit
//	save-cookies
//	save-session
//	server-stat-of
//
// In addition, options listed in the Input File subsection are available, except for following options: checksum, index-out, out, pause and select-file.
// With the log option, you can dynamically start logging or change log file.
// To stop logging, specify an empty string("") as the parameter value.
// Note that log file is always opened in append mode.
// This method returns OK for success.
func (c *client) ChangeGlobalOption(options Option) (ok string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, options)
	err = c.Call(aria2ChangeGlobalOption, params, &ok)
	return
}

// `aria2.getGlobalStat([secret])`
// This method returns global statistics such as the overall download and upload speeds.
// The response is a struct and contains the following keys. Values are strings.
//
//		downloadSpeed      Overall download speed (byte/sec).
//		uploadSpeed        Overall upload speed(byte/sec).
//		numActive          The number of active downloads.
//		numWaiting         The number of waiting downloads.
//		numStopped         The number of stopped downloads in the current session.
//	                    This value is capped by the --max-download-result option.
//		numStoppedTotal    The number of stopped downloads in the current session and not capped by the --max-download-result option.
func (c *client) GetGlobalStat() (info GlobalStatInfo, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2GetGlobalStat, params, &info)
	return
}

// `aria2.purgeDownloadResult([secret])`
// This method purges completed/error/removed downloads to free memory.
// This method returns OK.
func (c *client) PurgeDownloadResult() (ok string, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2PurgeDownloadResult, params, &ok)
	return
}

// `aria2.removeDownloadResult([secret, ]gid)`
// This method removes a completed/error/removed download denoted by gid from memory.
// This method returns OK for success.
func (c *client) RemoveDownloadResult(gid string) (ok string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2RemoveDownloadResult, params, &ok)
	return
}

// `aria2.getVersion([secret])`
// This method returns the version of aria2 and the list of enabled features.
// The response is a struct and contains following keys.
//
//	version            Version number of aria2 as a string.
//	enabledFeatures    List of enabled features. Each feature is given as a string.
func (c *client) GetVersion() (info VersionInfo, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2GetVersion, params, &info)
	return
}

// `aria2.getSessionInfo([secret])`
// This method returns session information.
// The response is a struct and contains following key.
//
//	sessionId    Session ID, which is generated each time when aria2 is invoked.
func (c *client) GetSessionInfo() (info SessionInfo, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2GetSessionInfo, params, &info)
	return
}

// `aria2.shutdown([secret])`
// This method shutdowns aria2.
// This method returns OK.
func (c *client) Shutdown() (ok string, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2Shutdown, params, &ok)
	return
}

// `aria2.forceShutdown([secret])`
// This method shuts down aria2().
// This method behaves like :func:'aria2.shutdown` without performing any actions which take time, such as contacting BitTorrent trackers to unregister downloads first.
// This method returns OK.
func (c *client) ForceShutdown() (ok string, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2ForceShutdown, params, &ok)
	return
}

// `aria2.saveSession([secret])`
// This method saves the current session to a file specified by the --save-session option.
// This method returns OK if it succeeds.
func (c *client) SaveSession() (ok string, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2SaveSession, params, &ok)
	return
}

// `system.multicall(methods)`
// This methods encapsulates multiple method calls in a single request.
// methods is an array of structs.
// The structs contain two keys: methodName and params.
// methodName is the method name to call and params is array containing parameters to the method call.
// This method returns an array of responses.
// The elements will be either a one-item array containing the return value of the method call or a struct of fault element if an encapsulated method call fails.
func (c *client) Multicall(methods []Method) (r []interface{}, err error) {
	if len(methods) == 0 {
		err = errInvalidParameter
		return
	}
	err = c.Call(aria2Multicall, []interface{}{methods}, &r)
	return
}

// `system.listMethods()`
// This method returns the all available RPC methods in an array of string.
// Unlike other methods, this method does not require secret token.
// This is safe because this method just returns the available method names.
func (c *client) ListMethods() (methods []string, err error) {
	err = c.Call(aria2ListMethods, []string{}, &methods)
	return
}
