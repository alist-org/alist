# Description
Command line tool for [gowebdav](https://github.com/studio-b12/gowebdav) library.

# Prerequisites
## Software
* **OS**: all, which are supported by `Golang`
* **Golang**: version 1.x
* **Git**: version 2.14.2 at higher (required to install via `go get`)

# Install
```sh
go get -u github.com/studio-b12/gowebdav/cmd/gowebdav
```

# Usage
It is recommended to set following environment variables to improve your experience with this tool:
* `ROOT` is an URL of target WebDAV server (e.g. `https://webdav.mydomain.me/user_root_folder`)
* `USER` is a login to connect to specified server (e.g. `user`)
* `PASSWORD` is a password to connect to specified server (e.g. `p@s$w0rD`)

In following examples we suppose that:
* environment variable `ROOT` is set to `https://webdav.mydomain.me/ufolder`
* environment variable `USER` is set to `user`
* environment variable `PASSWORD` is set `p@s$w0rD`
* folder `/ufolder/temp` exists on the server
* file `/ufolder/temp/file.txt` exists on the server
* file `/ufolder/temp/document.rtf` exists on the server
* file `/tmp/webdav/to_upload.txt` exists on the local machine
* folder `/tmp/webdav/` is used to download files from the server

## Examples

#### Get content of specified folder
```sh
gowebdav -X LS temp
```

#### Get info about file/folder
```sh
gowebdav -X STAT temp
gowebdav -X STAT temp/file.txt
```

#### Create folder on the remote server
```sh
gowebdav -X MKDIR temp2
gowebdav -X MKDIRALL all/folders/which-you-want/to_create
```

#### Download file
```sh
gowebdav -X GET temp/document.rtf /tmp/webdav/document.rtf
```

You may do not specify target local path, in this case file will be downloaded to the current folder with the

#### Upload file
```sh
gowebdav -X PUT temp/uploaded.txt /tmp/webdav/to_upload.txt
```

#### Move file on the remote server
```sh
gowebdav -X MV temp/file.txt temp/moved_file.txt
```

#### Copy file to another location
```sh
gowebdav -X MV temp/file.txt temp/file-copy.txt
```

#### Delete file from the remote server
```sh
gowebdav -X DEL temp/file.txt
```

# Wrapper script

You can create wrapper script for your server (via `$EDITOR ./dav && chmod a+x ./dav`) and add following content to it:
```sh
#!/bin/sh

ROOT="https://my.dav.server/" \
USER="foo" \
PASSWORD="$(pass dav/foo@my.dav.server)" \
gowebdav $@
```

It allows you to use [pass](https://www.passwordstore.org/ "the standard unix password manager") or similar tools to retrieve the password.

## Examples

Using the `dav` wrapper:

```sh
$ ./dav -X LS /

$ echo hi dav! > hello && ./dav -X PUT /hello
$ ./dav -X STAT /hello
$ ./dav -X PUT /hello_dav hello
$ ./dav -X GET /hello_dav
$ ./dav -X GET /hello_dav hello.txt
```