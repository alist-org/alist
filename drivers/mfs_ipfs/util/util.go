package util

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"sync"

	ipldformat "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-libipfs/files"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-mfs"
	pinningservice "github.com/ipfs/go-pinning-service-http-client"
	iface "github.com/ipfs/interface-go-ipfs-core"
	ifacepath "github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/ipfs/kubo/config"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/coreapi"
	"github.com/ipfs/kubo/plugin/loader"
	"github.com/ipfs/kubo/repo"
	"github.com/ipfs/kubo/repo/fsrepo"
)

type MfsAPI struct {
	cid   *string
	mroot *mfs.Root
}
type nodeObj struct {
	Id    string
	Name  string
	Size  int64
	Isdir bool
}

var DefaultPath = ""
var buildCount = -1
var buildLock sync.Mutex
var closeFunc func() error
var nodeApi iface.CoreAPI
var plugins = false
var repopath = ""

func buildInit() error {
	buildLock.Lock()
	defer buildLock.Unlock()
	if buildCount >= 0 {
		buildCount++
		return nil
	}
	buildCount = -1
	nodeApi = nil
	var err error
	if !plugins {
		var pluginload *loader.PluginLoader
		if pluginload, err = loader.NewPluginLoader(""); err == nil {
			if err = pluginload.Initialize(); err == nil {
				if err = pluginload.Inject(); err == nil {
					plugins = true
				}
			}
		}
	}
	newcfg := &config.Config{}
	if err == nil {
		newcfg, err = config.Init(io.Discard, 2048)
	}
	if err == nil {
		repopath = path.Join(DefaultPath, "ipfs_"+newcfg.Identity.PeerID)
		defer func() {
			r := recover()
			if err != nil || r != nil {
				os.RemoveAll(repopath)
				repopath = ""
			}
			if r != nil {
				panic(r)
			}
		}()
		err = fsrepo.Init(repopath, newcfg)
	}
	var newrepo repo.Repo
	if err == nil {
		newrepo, err = fsrepo.Open(repopath)
	}
	newnode := &core.IpfsNode{}
	if err == nil {
		closeFunc = newrepo.Close
		defer func() {
			r := recover()
			if err != nil || r != nil {
				closeFunc()
				closeFunc = nil
			}
			if r != nil {
				panic(r)
			}
		}()
		newnode, err = core.NewNode(context.Background(), &core.BuildCfg{Online: true, Repo: newrepo})
	}
	var newapi iface.CoreAPI
	if err == nil {
		closeFunc = newnode.Close
		newapi, err = coreapi.NewCoreAPI(newnode)
	}
	if err == nil {
		nodeApi = newapi
		buildCount = 1
	}
	return err
}
func buildDrop() {
	buildLock.Lock()
	defer buildLock.Unlock()
	if buildCount--; buildCount <= 0 {
		buildCount = -1
		nodeApi = nil
		defer os.RemoveAll(repopath)
		repopath = ""
		closeFunc()
		closeFunc = nil
	}
}
func NewMfs(cid *string) (*MfsAPI, error) {
	err := buildInit()
	if err != nil {
		return nil, err
	}
	defer func() {
		r := recover()
		if err != nil || r != nil {
			buildDrop()
		}
		if r != nil {
			panic(r)
		}
	}()
	ctx := context.Background()
	var rnode ipldformat.Node
	if err == nil {
		rnode, err = nodeApi.ResolveNode(ctx, ifacepath.New(*cid))
	}
	pnode := &merkledag.ProtoNode{}
	if err == nil {
		ok := true
		if pnode, ok = rnode.(*merkledag.ProtoNode); !ok {
			err = fmt.Errorf(reflect.TypeOf(rnode).String())
		}
	}
	mroot := &mfs.Root{}
	if err == nil {
		mroot, err = mfs.NewRoot(ctx, nodeApi.Dag(), pnode, nil)
	}
	if err != nil {
		return nil, err
	}
	return &MfsAPI{cid: cid, mroot: mroot}, err
}
func (mapi *MfsAPI) Close() error {
	defer buildDrop()
	return mapi.mroot.Close()
}
func (mapi *MfsAPI) List(spath string) ([]nodeObj, error) {
	pnode, err := mfs.Lookup(mapi.mroot, spath)
	dnode, ok := pnode.(*mfs.Directory)
	if err == nil && !ok {
		err = fmt.Errorf(reflect.TypeOf(pnode).String())
	}
	nl := []mfs.NodeListing{}
	ol := []nodeObj{}
	if err == nil {
		nl, err = dnode.List(context.Background())
	}
	if err == nil {
		for _, n := range nl {
			ol = append(ol, nodeObj{
				Id:    n.Hash,
				Name:  n.Name,
				Size:  n.Size,
				Isdir: n.Type == int(mfs.TDir),
			})
		}
	}
	return ol, err
}
func (mapi *MfsAPI) todo() {
	files.NewReaderFile(nil)
	pinningservice.NewClient("", "")
}
