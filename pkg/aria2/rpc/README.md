# PACKAGE DOCUMENTATION

**package rpc**
    
    import "github.com/matzoe/argo/rpc"



## FUNCTIONS

```
func Call(address, method string, params, reply interface{}) error
```

## TYPES

```
type Client struct {
    // contains filtered or unexported fields
}
```

```
func New(uri string) *Client
```

```
func (id *Client) AddMetalink(uri string, options ...interface{}) (gid string, err error)
```
`aria2.addMetalink(metalink[, options[, position]])` This method adds Metalink download by uploading ".metalink" file. `metalink` is of type base64 which contains Base64-encoded ".metalink" file. `options` is of type struct and its members are a pair of option name and value. See Options below for more details. If `position` is given as an integer starting from 0, the new download is inserted at `position` in the
waiting queue. If `position` is not given or `position` is larger than the size of the queue, it is appended at the end of the queue. This method returns array of GID of registered download. If `--rpc-save-upload-metadata` is true, the uploaded data is saved as a file named hex string of SHA-1 hash of data plus ".metalink" in the directory specified by `--dir` option. The example of filename is 0a3893293e27ac0490424c06de4d09242215f0a6.metalink. If same file already exists, it is overwritten. If the file cannot be saved successfully or `--rpc-save-upload-metadata` is false, the downloads added by this method are not saved by `--save-session`.

```
func (id *Client) AddTorrent(filename string, options ...interface{}) (gid string, err error)
```
`aria2.addTorrent(torrent[, uris[, options[, position]]])` This method adds BitTorrent download by uploading ".torrent" file. If you want to add BitTorrent Magnet URI, use `aria2.addUri()` method instead. torrent is of type base64 which contains Base64-encoded ".torrent" file. `uris` is of type array and its element is URI which is of type string. `uris` is used for Web-seeding. For single file torrents, URI can be a complete URI pointing to the resource or if URI ends with /, name in torrent file is added. For multi-file torrents, name and path in torrent are added to form a URI for each file. options is of type struct and its members are
a pair of option name and value. See Options below for more details. If `position` is given as an integer starting from 0, the new download is inserted at `position` in the waiting queue. If `position` is not given or `position` is larger than the size of the queue, it is appended at the end of the queue. This method returns GID of registered download. If `--rpc-save-upload-metadata` is true, the uploaded data is saved as a file named hex string of SHA-1 hash of data plus ".torrent" in the
directory specified by `--dir` option. The example of filename is 0a3893293e27ac0490424c06de4d09242215f0a6.torrent. If same file already exists, it is overwritten. If the file cannot be saved successfully or `--rpc-save-upload-metadata` is false, the downloads added by this method are not saved by -`-save-session`.

```
func (id *Client) AddUri(uri string, options ...interface{}) (gid string, err error)
```

`aria2.addUri(uris[, options[, position]])` This method adds new HTTP(S)/FTP/BitTorrent Magnet URI. `uris` is of type array and its element is URI which is of type string. For BitTorrent Magnet URI, `uris` must have only one element and it should be BitTorrent Magnet URI. URIs in uris must point to the same file. If you mix other URIs which point to another file, aria2 does not complain but download may
fail. `options` is of type struct and its members are a pair of option name and value. See Options below for more details. If `position` is given as an integer starting from 0, the new download is inserted at position in the waiting queue. If `position` is not given or `position` is larger than the size of the queue, it is appended at the end of the queue. This method returns GID of registered download.

```
func (id *Client) ChangeGlobalOption(options map[string]interface{}) (g string, err error)
```

`aria2.changeGlobalOption(options)` This method changes global options dynamically. `options` is of type struct. The following `options` are available:

    download-result
    log
    log-level
    max-concurrent-downloads
    max-download-result
    max-overall-download-limit
    max-overall-upload-limit
    save-cookies
    save-session
    server-stat-of

In addition to them, options listed in Input File subsection are available, except for following options: `checksum`, `index-out`, `out`, `pause` and `select-file`. Using `log` option, you can dynamically start logging or change log file. To stop logging, give empty string("") as a parameter value. Note that log file is always opened in append mode. This method returns OK for success.

```
func (id *Client) ChangeOption(gid string, options map[string]interface{}) (g string, err error)
```

`aria2.changeOption(gid, options)` This method changes options of the download denoted by `gid` dynamically. `gid` is of type string. `options` is of type struct. The following `options` are available for active downloads:

    bt-max-peers
    bt-request-peer-speed-limit
    bt-remove-unselected-file
    force-save
    max-download-limit
    max-upload-limit

For waiting or paused downloads, in addition to the above options, options listed in Input File subsection are available, except for following options: dry-run, metalink-base-uri, parameterized-uri, pause, piece-length and rpc-save-upload-metadata option. This method returns OK for success.

```
func (id *Client) ChangePosition(gid string, pos int, how string) (p int, err error)
```

`aria2.changePosition(gid, pos, how)` This method changes the position of the download denoted by `gid`. `pos` is of type integer. `how` is of type string. If `how` is `POS_SET`, it moves the download to a position relative to the beginning of the queue. If `how` is `POS_CUR`, it moves the download to a position relative to the current position. If `how` is `POS_END`, it moves the download to a position relative to the end of the queue. If the destination position is less than 0 or beyond the end
of the queue, it moves the download to the beginning or the end of the queue respectively. The response is of type integer and it is the destination position.

```
func (id *Client) ChangeUri(gid string, fileindex int, delUris []string, addUris []string, position ...int) (p []int, err error)
```

`aria2.changeUri(gid, fileIndex, delUris, addUris[, position])` This method removes URIs in `delUris` from and appends URIs in `addUris` to download denoted by gid. `delUris` and `addUris` are list of string. A download can contain multiple files and URIs are attached to each file. `fileIndex` is used to select which file to remove/attach given URIs. `fileIndex` is 1-based. `position` is used to specify where URIs are inserted in the existing waiting URI list. `position` is 0-based. When
`position` is omitted, URIs are appended to the back of the list. This method first execute removal and then addition. `position` is the `position` after URIs are removed, not the `position` when this method is called. When removing URI, if same URIs exist in download, only one of them is removed for each URI in delUris. In other words, there are three URIs http://example.org/aria2 and you want remove them all, you
have to specify (at least) 3 http://example.org/aria2 in delUris. This method returns a list which contains 2 integers. The first integer is the number of URIs deleted. The second integer is the number of URIs added.

```
func (id *Client) ForcePause(gid string) (g string, err error)
```

`aria2.forcePause(pid)` This method pauses the download denoted by `gid`. This method behaves just like aria2.pause() except that this method pauses download without any action which takes time such as contacting BitTorrent tracker.

```
func (id *Client) ForcePauseAll() (g string, err error)
```

`aria2.forcePauseAll()` This method is equal to calling `aria2.forcePause()` for every active/waiting download. This methods returns OK for success.

```
func (id *Client) ForceRemove(gid string) (g string, err error)
```

`aria2.forceRemove(gid)` This method removes the download denoted by `gid`. This method behaves just like aria2.remove() except that this method removes download without any action which takes time such as contacting BitTorrent tracker.

```
func (id *Client) ForceShutdown() (g string, err error)
```

`aria2.forceShutdown()` This method shutdowns aria2. This method behaves like `aria2.shutdown()` except that any actions which takes time such as contacting BitTorrent tracker are skipped. This method returns OK. 

```
func (id *Client) GetFiles(gid string) (m map[string]interface{}, err error)
```

`aria2.getFiles(gid)` This method returns file list of the download denoted by `gid`. `gid` is of type string.

```
func (id *Client) GetGlobalOption() (m map[string]interface{}, err error)
```

`aria2.getGlobalOption()` This method returns global options. The response is of type struct. Its key is the name of option. The value type is string. Note that this method does not return options which have no default value and have not been set by the command-line options, configuration files or RPC methods. Because global options are used as a template for the options of newly added download, the response contains
keys returned by `aria2.getOption()` method.

```
func (id *Client) GetGlobalStat() (m map[string]interface{}, err error)
```

`aria2.getGlobalStat()` This method returns global statistics such as overall download and upload speed.

```
func (id *Client) GetOption(gid string) (m map[string]interface{}, err error)
```

`aria2.getOption(gid)` This method returns options of the download denoted by `gid`. The response is of type struct. Its key is the name of option. The value type is string. Note that this method does not return options which have no default value and have not been set by the command-line options, configuration files or RPC methods.

```
func (id *Client) GetPeers(gid string) (m []map[string]interface{}, err error)
```

`aria2.getPeers(gid)` This method returns peer list of the download denoted by `gid`. `gid` is of type string. This method is for BitTorrent only.

```
func (id *Client) GetServers(gid string) (m []map[string]interface{}, err error)
```

`aria2.getServers(gid)` This method returns currently connected HTTP(S)/FTP servers of the download denoted by `gid`. `gid` is of type string.

```
func (id *Client) GetSessionInfo() (m map[string]interface{}, err error)
```

`aria2.getSessionInfo()` This method returns session information.

```
func (id *Client) GetUris(gid string) (m map[string]interface{}, err error)
```

`aria2.getUris(gid)` This method returns URIs used in the download denoted by `gid`. `gid` is of type string.

```
func (id *Client) GetVersion() (m map[string]interface{}, err error)
```

`aria2.getVersion()` This method returns version of the program and the list of enabled features.

```
func (id *Client) Multicall(methods []map[string]interface{}) (r []interface{}, err error)
```

`system.multicall(methods)` This method encapsulates multiple method calls in a single request. `methods` is of type array and its element is struct. The struct contains two keys: `methodName` and `params`. `methodName` is the method name to call and `params` is array containing parameters to the method. This method returns array of responses. The element of array will either be a one-item array containing the return value of each method call or struct of fault element if an encapsulated method call fails.

```
func (id *Client) Pause(gid string) (g string, err error)
```

`aria2.pause(gid)` This method pauses the download denoted by `gid`. `gid` is of type string. The status of paused download becomes paused. If the download is active, the download is placed on the first position of waiting queue. As long as the status is paused, the download is not started. To change status to waiting, use `aria2.unpause()` method. This method returns GID of paused download.

```
func (id *Client) PauseAll() (g string, err error)
```

`aria2.pauseAll()` This method is equal to calling `aria2.pause()` for every active/waiting download. This methods returns OK for success.

```
func (id *Client) PurgeDownloadResult() (g string, err error)
```

`aria2.purgeDownloadResult()` This method purges completed/error/removed downloads to free memory. This method returns OK.

```
func (id *Client) Remove(gid string) (g string, err error)
```

`aria2.remove(gid)` This method removes the download denoted by gid. `gid` is of type string. If specified download is in progress, it is stopped at first. The status of removed download becomes removed. This method returns GID of removed download.

```
func (id *Client) RemoveDownloadResult(gid string) (g string, err error)
```

`aria2.removeDownloadResult(gid)` This method removes completed/error/removed download denoted by `gid` from memory. This method returns OK for success.

```
func (id *Client) Shutdown() (g string, err error)
```

`aria2.shutdown()` This method shutdowns aria2. This method returns OK.

```
func (id *Client) TellActive(keys ...string) (m []map[string]interface{}, err error)
```

`aria2.tellActive([keys])` This method returns the list of active downloads. The response is of type array and its element is the same struct returned by `aria2.tellStatus()` method. For `keys` parameter, please refer to `aria2.tellStatus()` method.

```
func (id *Client) TellStatus(gid string, keys ...string) (m map[string]interface{}, err error)
```

`aria2.tellStatus(gid[, keys])` This method returns download progress of the download denoted by `gid`. `gid` is of type string. `keys` is array of string. If it is specified, the response contains only keys in `keys` array. If `keys` is empty or not specified, the response contains all keys. This is useful when you just want specific keys and avoid unnecessary transfers. For example, `aria2.tellStatus("2089b05ecca3d829", ["gid", "status"])` returns `gid` and `status` key.

```
func (id *Client) TellStopped(offset, num int, keys ...string) (m []map[string]interface{}, err error)
```

`aria2.tellStopped(offset, num[, keys])` This method returns the list of stopped download. `offset` is of type integer and specifies the `offset` from the oldest download. `num` is of type integer and specifies the number of downloads to be returned. For keys parameter, please refer to `aria2.tellStatus()` method. `offset` and `num` have the same semantics as `aria2.tellWaiting()` method. The response is of type array and its element is the same struct returned by `aria2.tellStatus()` method.

```
func (id *Client) TellWaiting(offset, num int, keys ...string) (m []map[string]interface{}, err error)
```
`aria2.tellWaiting(offset, num[, keys])` This method returns the list of waiting download, including paused downloads. `offset` is of type integer and specifies the `offset` from the download waiting at the front. num is of type integer and specifies the number of downloads to be returned. For keys parameter, please refer to aria2.tellStatus() method. If `offset` is a positive integer, this method returns downloads
in the range of `[offset, offset + num)`. `offset` can be a negative integer. `offset == -1` points last download in the waiting queue and `offset == -2` points the download before the last download, and so on. The downloads in the response are in reversed order. For example, imagine that three downloads "A","B" and "C" are waiting in this order.

    aria2.tellWaiting(0, 1) returns ["A"].
    aria2.tellWaiting(1, 2) returns ["B", "C"].
    aria2.tellWaiting(-1, 2) returns ["C", "B"].

The response is of type array and its element is the same struct returned by `aria2.tellStatus()` method.

```
func (id *Client) Unpause(gid string) (g string, err error)
```

`aria2.unpause(gid)` This method changes the status of the download denoted by `gid` from paused to waiting. This makes the download eligible to restart. `gid` is of type string. This method returns GID of unpaused download.

```
func (id *Client) UnpauseAll() (g string, err error)
```

`aria2.unpauseAll()` This method is equal to calling `aria2.unpause()` for every active/waiting download. This methods returns OK for success.
