# GoWebDAV

[![Build Status](https://travis-ci.org/studio-b12/gowebdav.svg?branch=master)](https://travis-ci.org/studio-b12/gowebdav)
[![GoDoc](https://godoc.org/github.com/studio-b12/gowebdav?status.svg)](https://godoc.org/github.com/studio-b12/gowebdav)
[![Go Report Card](https://goreportcard.com/badge/github.com/studio-b12/gowebdav)](https://goreportcard.com/report/github.com/studio-b12/gowebdav)

A golang WebDAV client library.

## Main features
`gowebdav` library allows to perform following actions on the remote WebDAV server:
* [create path](#create-path-on-a-webdav-server)
* [get files list](#get-files-list)
* [download file](#download-file-to-byte-array)
* [upload file](#upload-file-from-byte-array)
* [get information about specified file/folder](#get-information-about-specified-filefolder)
* [move file to another location](#move-file-to-another-location)
* [copy file to another location](#copy-file-to-another-location)
* [delete file](#delete-file)

## Usage

First of all you should create `Client` instance using `NewClient()` function:

```go
root := "https://webdav.mydomain.me"
user := "user"
password := "password"

c := gowebdav.NewClient(root, user, password)
```

After you can use this `Client` to perform actions, described below.

**NOTICE:** we will not check errors in examples, to focus you on the `gowebdav` library's code, but you should do it in your code!

### Create path on a WebDAV server
```go
err := c.Mkdir("folder", 0644)
```
In case you want to create several folders you can use `c.MkdirAll()`:
```go
err := c.MkdirAll("folder/subfolder/subfolder2", 0644)
```

### Get files list
```go
files, _ := c.ReadDir("folder/subfolder")
for _, file := range files {
    //notice that [file] has os.FileInfo type
    fmt.Println(file.Name())
}
```

### Download file to byte array
```go
webdavFilePath := "folder/subfolder/file.txt"
localFilePath := "/tmp/webdav/file.txt"

bytes, _ := c.Read(webdavFilePath)
ioutil.WriteFile(localFilePath, bytes, 0644)
```

### Download file via reader
Also you can use `c.ReadStream()` method:
```go
webdavFilePath := "folder/subfolder/file.txt"
localFilePath := "/tmp/webdav/file.txt"

reader, _ := c.ReadStream(webdavFilePath)

file, _ := os.Create(localFilePath)
defer file.Close()

io.Copy(file, reader)
```

### Upload file from byte array
```go
webdavFilePath := "folder/subfolder/file.txt"
localFilePath := "/tmp/webdav/file.txt"

bytes, _ := ioutil.ReadFile(localFilePath)

c.Write(webdavFilePath, bytes, 0644)
```

### Upload file via writer
```go
webdavFilePath := "folder/subfolder/file.txt"
localFilePath := "/tmp/webdav/file.txt"

file, _ := os.Open(localFilePath)
defer file.Close()

c.WriteStream(webdavFilePath, file, 0644)
```

### Get information about specified file/folder
```go
webdavFilePath := "folder/subfolder/file.txt"

info := c.Stat(webdavFilePath)
//notice that [info] has os.FileInfo type
fmt.Println(info)
```

### Move file to another location
```go
oldPath := "folder/subfolder/file.txt"
newPath := "folder/subfolder/moved.txt"
isOverwrite := true

c.Rename(oldPath, newPath, isOverwrite)
```

### Copy file to another location
```go
oldPath := "folder/subfolder/file.txt"
newPath := "folder/subfolder/file-copy.txt"
isOverwrite := true

c.Copy(oldPath, newPath, isOverwrite)
```

### Delete file
```go
webdavFilePath := "folder/subfolder/file.txt"

c.Remove(webdavFilePath)
```

## Links

More details about WebDAV server you can read from following resources:

* [RFC 4918 - HTTP Extensions for Web Distributed Authoring and Versioning (WebDAV)](https://tools.ietf.org/html/rfc4918)
* [RFC 5689 - Extended MKCOL for Web Distributed Authoring and Versioning (WebDAV)](https://tools.ietf.org/html/rfc5689)
* [RFC 2616 - HTTP/1.1 Status Code Definitions](http://www.w3.org/Protocols/rfc2616/rfc2616-sec10.html "HTTP/1.1 Status Code Definitions")
* [WebDav: Next Generation Collaborative Web Authoring By Lisa Dusseaul](https://books.google.de/books?isbn=0130652083 "WebDav: Next Generation Collaborative Web Authoring By Lisa Dusseault")

**NOTICE**: RFC 2518 is obsoleted by RFC 4918 in June 2007

## Contributing
All contributing are welcome. If you have any suggestions or find some bug - please create an Issue to let us make this project better. We appreciate your help!

## License
This library is distributed under the BSD 3-Clause license found in the [LICENSE](https://github.com/studio-b12/gowebdav/blob/master/LICENSE) file.
## API

`import "github.com/studio-b12/gowebdav"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)
* [Examples](#pkg-examples)
* [Subdirectories](#pkg-subdirectories)

### <a name="pkg-overview">Overview</a>
Package gowebdav is a WebDAV client library with a command line tool
included.

### <a name="pkg-index">Index</a>
* [func FixSlash(s string) string](#FixSlash)
* [func FixSlashes(s string) string](#FixSlashes)
* [func Join(path0 string, path1 string) string](#Join)
* [func PathEscape(path string) string](#PathEscape)
* [func ReadConfig(uri, netrc string) (string, string)](#ReadConfig)
* [func String(r io.Reader) string](#String)
* [type Authenticator](#Authenticator)
* [type BasicAuth](#BasicAuth)
  * [func (b *BasicAuth) Authorize(req *http.Request, method string, path string)](#BasicAuth.Authorize)
  * [func (b *BasicAuth) Pass() string](#BasicAuth.Pass)
  * [func (b *BasicAuth) Type() string](#BasicAuth.Type)
  * [func (b *BasicAuth) User() string](#BasicAuth.User)
* [type Client](#Client)
  * [func NewClient(uri, user, pw string) *Client](#NewClient)
  * [func (c *Client) Connect() error](#Client.Connect)
  * [func (c *Client) Copy(oldpath, newpath string, overwrite bool) error](#Client.Copy)
  * [func (c *Client) Mkdir(path string, _ os.FileMode) error](#Client.Mkdir)
  * [func (c *Client) MkdirAll(path string, _ os.FileMode) error](#Client.MkdirAll)
  * [func (c *Client) Read(path string) ([]byte, error)](#Client.Read)
  * [func (c *Client) ReadDir(path string) ([]os.FileInfo, error)](#Client.ReadDir)
  * [func (c *Client) ReadStream(path string) (io.ReadCloser, error)](#Client.ReadStream)
  * [func (c *Client) ReadStreamRange(path string, offset, length int64) (io.ReadCloser, error)](#Client.ReadStreamRange)
  * [func (c *Client) Remove(path string) error](#Client.Remove)
  * [func (c *Client) RemoveAll(path string) error](#Client.RemoveAll)
  * [func (c *Client) Rename(oldpath, newpath string, overwrite bool) error](#Client.Rename)
  * [func (c *Client) SetHeader(key, value string)](#Client.SetHeader)
  * [func (c *Client) SetInterceptor(interceptor func(method string, rq *http.Request))](#Client.SetInterceptor)
  * [func (c *Client) SetTimeout(timeout time.Duration)](#Client.SetTimeout)
  * [func (c *Client) SetTransport(transport http.RoundTripper)](#Client.SetTransport)
  * [func (c *Client) Stat(path string) (os.FileInfo, error)](#Client.Stat)
  * [func (c *Client) Write(path string, data []byte, _ os.FileMode) error](#Client.Write)
  * [func (c *Client) WriteStream(path string, stream io.Reader, _ os.FileMode) error](#Client.WriteStream)
* [type DigestAuth](#DigestAuth)
  * [func (d *DigestAuth) Authorize(req *http.Request, method string, path string)](#DigestAuth.Authorize)
  * [func (d *DigestAuth) Pass() string](#DigestAuth.Pass)
  * [func (d *DigestAuth) Type() string](#DigestAuth.Type)
  * [func (d *DigestAuth) User() string](#DigestAuth.User)
* [type File](#File)
  * [func (f File) ContentType() string](#File.ContentType)
  * [func (f File) ETag() string](#File.ETag)
  * [func (f File) IsDir() bool](#File.IsDir)
  * [func (f File) ModTime() time.Time](#File.ModTime)
  * [func (f File) Mode() os.FileMode](#File.Mode)
  * [func (f File) Name() string](#File.Name)
  * [func (f File) Path() string](#File.Path)
  * [func (f File) Size() int64](#File.Size)
  * [func (f File) String() string](#File.String)
  * [func (f File) Sys() interface{}](#File.Sys)
* [type NoAuth](#NoAuth)
  * [func (n *NoAuth) Authorize(req *http.Request, method string, path string)](#NoAuth.Authorize)
  * [func (n *NoAuth) Pass() string](#NoAuth.Pass)
  * [func (n *NoAuth) Type() string](#NoAuth.Type)
  * [func (n *NoAuth) User() string](#NoAuth.User)

##### <a name="pkg-examples">Examples</a>
* [PathEscape](#example_PathEscape)

##### <a name="pkg-files">Package files</a>
[basicAuth.go](https://github.com/studio-b12/gowebdav/blob/master/basicAuth.go) [client.go](https://github.com/studio-b12/gowebdav/blob/master/client.go) [digestAuth.go](https://github.com/studio-b12/gowebdav/blob/master/digestAuth.go) [doc.go](https://github.com/studio-b12/gowebdav/blob/master/doc.go) [file.go](https://github.com/studio-b12/gowebdav/blob/master/file.go) [netrc.go](https://github.com/studio-b12/gowebdav/blob/master/netrc.go) [requests.go](https://github.com/studio-b12/gowebdav/blob/master/requests.go) [utils.go](https://github.com/studio-b12/gowebdav/blob/master/utils.go) 

### <a name="FixSlash">func</a> [FixSlash](https://github.com/studio-b12/gowebdav/blob/master/utils.go?s=707:737#L45)
``` go
func FixSlash(s string) string
```
FixSlash appends a trailing / to our string

### <a name="FixSlashes">func</a> [FixSlashes](https://github.com/studio-b12/gowebdav/blob/master/utils.go?s=859:891#L53)
``` go
func FixSlashes(s string) string
```
FixSlashes appends and prepends a / if they are missing

### <a name="Join">func</a> [Join](https://github.com/studio-b12/gowebdav/blob/master/utils.go?s=992:1036#L62)
``` go
func Join(path0 string, path1 string) string
```
Join joins two paths

### <a name="PathEscape">func</a> [PathEscape](https://github.com/studio-b12/gowebdav/blob/master/utils.go?s=506:541#L36)
``` go
func PathEscape(path string) string
```
PathEscape escapes all segments of a given path

### <a name="ReadConfig">func</a> [ReadConfig](https://github.com/studio-b12/gowebdav/blob/master/netrc.go?s=428:479#L27)
``` go
func ReadConfig(uri, netrc string) (string, string)
```
ReadConfig reads login and password configuration from ~/.netrc
machine foo.com login username password 123456

### <a name="String">func</a> [String](https://github.com/studio-b12/gowebdav/blob/master/utils.go?s=1166:1197#L67)
``` go
func String(r io.Reader) string
```
String pulls a string out of our io.Reader

### <a name="Authenticator">type</a> [Authenticator](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=388:507#L29)
``` go
type Authenticator interface {
    Type() string
    User() string
    Pass() string
    Authorize(*http.Request, string, string)
}
```
Authenticator stub

### <a name="BasicAuth">type</a> [BasicAuth](https://github.com/studio-b12/gowebdav/blob/master/basicAuth.go?s=106:157#L9)
``` go
type BasicAuth struct {
    // contains filtered or unexported fields
}
```
BasicAuth structure holds our credentials

#### <a name="BasicAuth.Authorize">func</a> (\*BasicAuth) [Authorize](https://github.com/studio-b12/gowebdav/blob/master/basicAuth.go?s=473:549#L30)
``` go
func (b *BasicAuth) Authorize(req *http.Request, method string, path string)
```
Authorize the current request

#### <a name="BasicAuth.Pass">func</a> (\*BasicAuth) [Pass](https://github.com/studio-b12/gowebdav/blob/master/basicAuth.go?s=388:421#L25)
``` go
func (b *BasicAuth) Pass() string
```
Pass holds the BasicAuth password

#### <a name="BasicAuth.Type">func</a> (\*BasicAuth) [Type](https://github.com/studio-b12/gowebdav/blob/master/basicAuth.go?s=201:234#L15)
``` go
func (b *BasicAuth) Type() string
```
Type identifies the BasicAuthenticator

#### <a name="BasicAuth.User">func</a> (\*BasicAuth) [User](https://github.com/studio-b12/gowebdav/blob/master/basicAuth.go?s=297:330#L20)
``` go
func (b *BasicAuth) User() string
```
User holds the BasicAuth username

### <a name="Client">type</a> [Client](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=172:364#L18)
``` go
type Client struct {
    // contains filtered or unexported fields
}
```
Client defines our structure

#### <a name="NewClient">func</a> [NewClient](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=1019:1063#L62)
``` go
func NewClient(uri, user, pw string) *Client
```
NewClient creates a new instance of client

#### <a name="Client.Connect">func</a> (\*Client) [Connect](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=1843:1875#L87)
``` go
func (c *Client) Connect() error
```
Connect connects to our dav server

#### <a name="Client.Copy">func</a> (\*Client) [Copy](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=6702:6770#L313)
``` go
func (c *Client) Copy(oldpath, newpath string, overwrite bool) error
```
Copy copies a file from A to B

#### <a name="Client.Mkdir">func</a> (\*Client) [Mkdir](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=5793:5849#L272)
``` go
func (c *Client) Mkdir(path string, _ os.FileMode) error
```
Mkdir makes a directory

#### <a name="Client.MkdirAll">func</a> (\*Client) [MkdirAll](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=6028:6087#L283)
``` go
func (c *Client) MkdirAll(path string, _ os.FileMode) error
```
MkdirAll like mkdir -p, but for webdav

#### <a name="Client.Read">func</a> (\*Client) [Read](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=6876:6926#L318)
``` go
func (c *Client) Read(path string) ([]byte, error)
```
Read reads the contents of a remote file

#### <a name="Client.ReadDir">func</a> (\*Client) [ReadDir](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=2869:2929#L130)
``` go
func (c *Client) ReadDir(path string) ([]os.FileInfo, error)
```
ReadDir reads the contents of a remote directory

#### <a name="Client.ReadStream">func</a> (\*Client) [ReadStream](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=7237:7300#L336)
``` go
func (c *Client) ReadStream(path string) (io.ReadCloser, error)
```
ReadStream reads the stream for a given path

#### <a name="Client.ReadStreamRange">func</a> (\*Client) [ReadStreamRange](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=8049:8139#L358)
``` go
func (c *Client) ReadStreamRange(path string, offset, length int64) (io.ReadCloser, error)
```
ReadStreamRange reads the stream representing a subset of bytes for a given path,
utilizing HTTP Range Requests if the server supports it.
The range is expressed as offset from the start of the file and length, for example
offset=10, length=10 will return bytes 10 through 19.

If the server does not support partial content requests and returns full content instead,
this function will emulate the behavior by skipping `offset` bytes and limiting the result
to `length`.

#### <a name="Client.Remove">func</a> (\*Client) [Remove](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=5299:5341#L249)
``` go
func (c *Client) Remove(path string) error
```
Remove removes a remote file

#### <a name="Client.RemoveAll">func</a> (\*Client) [RemoveAll](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=5407:5452#L254)
``` go
func (c *Client) RemoveAll(path string) error
```
RemoveAll removes remote files

#### <a name="Client.Rename">func</a> (\*Client) [Rename](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=6536:6606#L308)
``` go
func (c *Client) Rename(oldpath, newpath string, overwrite bool) error
```
Rename moves a file from A to B

#### <a name="Client.SetHeader">func</a> (\*Client) [SetHeader](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=1235:1280#L67)
``` go
func (c *Client) SetHeader(key, value string)
```
SetHeader lets us set arbitrary headers for a given client

#### <a name="Client.SetInterceptor">func</a> (\*Client) [SetInterceptor](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=1387:1469#L72)
``` go
func (c *Client) SetInterceptor(interceptor func(method string, rq *http.Request))
```
SetInterceptor lets us set an arbitrary interceptor for a given client

#### <a name="Client.SetTimeout">func</a> (\*Client) [SetTimeout](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=1571:1621#L77)
``` go
func (c *Client) SetTimeout(timeout time.Duration)
```
SetTimeout exposes the ability to set a time limit for requests

#### <a name="Client.SetTransport">func</a> (\*Client) [SetTransport](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=1714:1772#L82)
``` go
func (c *Client) SetTransport(transport http.RoundTripper)
```
SetTransport exposes the ability to define custom transports

#### <a name="Client.Stat">func</a> (\*Client) [Stat](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=4255:4310#L197)
``` go
func (c *Client) Stat(path string) (os.FileInfo, error)
```
Stat returns the file stats for a specified path

#### <a name="Client.Write">func</a> (\*Client) [Write](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=9051:9120#L388)
``` go
func (c *Client) Write(path string, data []byte, _ os.FileMode) error
```
Write writes data to a given path

#### <a name="Client.WriteStream">func</a> (\*Client) [WriteStream](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=9476:9556#L411)
``` go
func (c *Client) WriteStream(path string, stream io.Reader, _ os.FileMode) error
```
WriteStream writes a stream

### <a name="DigestAuth">type</a> [DigestAuth](https://github.com/studio-b12/gowebdav/blob/master/digestAuth.go?s=157:254#L14)
``` go
type DigestAuth struct {
    // contains filtered or unexported fields
}
```
DigestAuth structure holds our credentials

#### <a name="DigestAuth.Authorize">func</a> (\*DigestAuth) [Authorize](https://github.com/studio-b12/gowebdav/blob/master/digestAuth.go?s=577:654#L36)
``` go
func (d *DigestAuth) Authorize(req *http.Request, method string, path string)
```
Authorize the current request

#### <a name="DigestAuth.Pass">func</a> (\*DigestAuth) [Pass](https://github.com/studio-b12/gowebdav/blob/master/digestAuth.go?s=491:525#L31)
``` go
func (d *DigestAuth) Pass() string
```
Pass holds the DigestAuth password

#### <a name="DigestAuth.Type">func</a> (\*DigestAuth) [Type](https://github.com/studio-b12/gowebdav/blob/master/digestAuth.go?s=299:333#L21)
``` go
func (d *DigestAuth) Type() string
```
Type identifies the DigestAuthenticator

#### <a name="DigestAuth.User">func</a> (\*DigestAuth) [User](https://github.com/studio-b12/gowebdav/blob/master/digestAuth.go?s=398:432#L26)
``` go
func (d *DigestAuth) User() string
```
User holds the DigestAuth username

### <a name="File">type</a> [File](https://github.com/studio-b12/gowebdav/blob/master/file.go?s=93:253#L10)
``` go
type File struct {
    // contains filtered or unexported fields
}
```
File is our structure for a given file

#### <a name="File.ContentType">func</a> (File) [ContentType](https://github.com/studio-b12/gowebdav/blob/master/file.go?s=476:510#L31)
``` go
func (f File) ContentType() string
```
ContentType returns the content type of a file

#### <a name="File.ETag">func</a> (File) [ETag](https://github.com/studio-b12/gowebdav/blob/master/file.go?s=929:956#L56)
``` go
func (f File) ETag() string
```
ETag returns the ETag of a file

#### <a name="File.IsDir">func</a> (File) [IsDir](https://github.com/studio-b12/gowebdav/blob/master/file.go?s=1035:1061#L61)
``` go
func (f File) IsDir() bool
```
IsDir let us see if a given file is a directory or not

#### <a name="File.ModTime">func</a> (File) [ModTime](https://github.com/studio-b12/gowebdav/blob/master/file.go?s=836:869#L51)
``` go
func (f File) ModTime() time.Time
```
ModTime returns the modified time of a file

#### <a name="File.Mode">func</a> (File) [Mode](https://github.com/studio-b12/gowebdav/blob/master/file.go?s=665:697#L41)
``` go
func (f File) Mode() os.FileMode
```
Mode will return the mode of a given file

#### <a name="File.Name">func</a> (File) [Name](https://github.com/studio-b12/gowebdav/blob/master/file.go?s=378:405#L26)
``` go
func (f File) Name() string
```
Name returns the name of a file

#### <a name="File.Path">func</a> (File) [Path](https://github.com/studio-b12/gowebdav/blob/master/file.go?s=295:322#L21)
``` go
func (f File) Path() string
```
Path returns the full path of a file

#### <a name="File.Size">func</a> (File) [Size](https://github.com/studio-b12/gowebdav/blob/master/file.go?s=573:599#L36)
``` go
func (f File) Size() int64
```
Size returns the size of a file

#### <a name="File.String">func</a> (File) [String](https://github.com/studio-b12/gowebdav/blob/master/file.go?s=1183:1212#L71)
``` go
func (f File) String() string
```
String lets us see file information

#### <a name="File.Sys">func</a> (File) [Sys](https://github.com/studio-b12/gowebdav/blob/master/file.go?s=1095:1126#L66)
``` go
func (f File) Sys() interface{}
```
Sys ????

### <a name="NoAuth">type</a> [NoAuth](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=551:599#L37)
``` go
type NoAuth struct {
    // contains filtered or unexported fields
}
```
NoAuth structure holds our credentials

#### <a name="NoAuth.Authorize">func</a> (\*NoAuth) [Authorize](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=894:967#L58)
``` go
func (n *NoAuth) Authorize(req *http.Request, method string, path string)
```
Authorize the current request

#### <a name="NoAuth.Pass">func</a> (\*NoAuth) [Pass](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=812:842#L53)
``` go
func (n *NoAuth) Pass() string
```
Pass returns the current password

#### <a name="NoAuth.Type">func</a> (\*NoAuth) [Type](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=638:668#L43)
``` go
func (n *NoAuth) Type() string
```
Type identifies the authenticator

#### <a name="NoAuth.User">func</a> (\*NoAuth) [User](https://github.com/studio-b12/gowebdav/blob/master/client.go?s=724:754#L48)
``` go
func (n *NoAuth) User() string
```
User returns the current user

- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
