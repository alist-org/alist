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
	"github.com/libp2p/go-libp2p/core/peer"
)

type MfsAPI struct {
	CID       *string
	PinID     *string
	lock      sync.RWMutex
	mroot     *mfs.Root
	pinclient *pinningservice.Client
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

func NewMfs(purl, ptoken string) (mapi *MfsAPI, err error) {
	buildLock.Lock()
	defer buildLock.Unlock()
	if buildCount >= 0 {
		buildCount++
		return &MfsAPI{
			pinclient: pinningservice.NewClient(purl, ptoken),
		}, nil
	}
	buildCount = -1
	nodeApi = nil
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
			if mapi == nil {
				os.RemoveAll(repopath)
				repopath = ""
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
			if mapi == nil {
				closeFunc()
				closeFunc = nil
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
		return &MfsAPI{
			pinclient: pinningservice.NewClient(purl, ptoken),
		}, nil
	}
	return
}
func (mapi *MfsAPI) Close() (err error) {
	mapi.lock.Lock()
	defer func() {
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
	}()
	if mapi.mroot != nil {
		err = mapi.mroot.Close()
		mapi.mroot = nil
	}
	return
}
func (mapi *MfsAPI) newRoot(force bool) (err error) {
	mapi.lock.Lock()
	defer mapi.lock.Unlock()
	if !force && mapi.mroot != nil {
		return
	}
	emptystr := ""
	pinid := &emptystr
	rootcid := &emptystr
	if ptr := mapi.PinID; ptr != nil {
		pinid = ptr
	}
	if ptr := mapi.CID; ptr != nil {
		rootcid = ptr
	}
	ctx := context.Background()
	if pinstatus, err := mapi.pinclient.GetStatusByID(ctx, *pinid); err == nil {
		if info, err := peer.AddrInfosFromP2pAddrs(pinstatus.GetDelegates()...); err == nil {
			for _, a := range info {
				nodeApi.Swarm().Connect(ctx, a)
			}
		}
		if pinstatus.GetStatus() == pinningservice.StatusPinned {
			*rootcid = pinstatus.GetPin().GetCid().String()
		}
	}
	var rnode ipldformat.Node
	rnode, err = nodeApi.ResolveNode(ctx, ifacepath.New(*rootcid))
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
	if err == nil {
		mapi.mroot = mroot
	}
	return
}
func (mapi *MfsAPI) List(spath string) (ol []nodeObj, err error) {
	if mapi.mroot == nil {
		if err = mapi.newRoot(false); err != nil {
			return
		}
	}
	mapi.lock.RLock()
	defer mapi.lock.RUnlock()
	pnode, err := mfs.Lookup(mapi.mroot, spath)
	dnode, ok := pnode.(*mfs.Directory)
	if err == nil && !ok {
		err = fmt.Errorf(reflect.TypeOf(pnode).String())
	}
	nl := []mfs.NodeListing{}
	if err == nil {
		nl, err = dnode.List(context.Background())
	}
	if err == nil {
		ol = []nodeObj{}
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
func (mapi *MfsAPI) Put(rc io.ReadCloser) (err error) {
	if mapi.mroot == nil {
		if err = mapi.newRoot(false); err != nil {
			return
		}
	}
	mapi.lock.RLock()
	defer mapi.lock.RUnlock()
	files.NewReaderFile(rc)
	return
}
