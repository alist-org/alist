package util

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
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
	cancel    context.CancelFunc
	ctx       context.Context
	lock      *sync.RWMutex
	mroot     *mfs.Root
	newcid    *cid.Cid
	pinclient *pinningservice.Client
	pinlock   *sync.Mutex
	pinning   []chan<- error
	queued    []chan<- error
	refcid    refresher
	refpinid  refresher
}
type NodeObj struct {
	Id    string
	Name  string
	Size  int64
	Isdir bool
}
type refresher interface {
	Get() (id string)
	Set(id string)
}
type emptyref struct{}

func (emptyref) Get() (id string) { return "" }
func (emptyref) Set(id string)    {}

var Ctx = context.Background()
var DefaultPath = ""
var buildCount = -1
var buildLock sync.Mutex
var closeFunc func() error
var nodeApi iface.CoreAPI
var plugins = false
var repopath = ""

func NewMfs(purl, ptoken string, rcid, rpinid refresher) (mapi *MfsAPI, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
		if mapi != nil {
			go mapi.runPin()
			go mapi.List("")
		}
	}()
	buildLock.Lock()
	defer buildLock.Unlock()
	if buildCount >= 0 {
		buildCount++
		mapi = &MfsAPI{
			lock:      &sync.RWMutex{},
			mroot:     nil,
			newcid:    nil,
			pinclient: pinningservice.NewClient(purl, ptoken),
			pinlock:   &sync.Mutex{},
			pinning:   []chan<- error{},
			queued:    []chan<- error{},
			refcid:    rcid,
			refpinid:  rpinid,
		}
		mapi.ctx, mapi.cancel = context.WithCancel(Ctx)
		if mapi.refcid == nil {
			mapi.refcid = emptyref{}
		}
		if mapi.refpinid == nil {
			mapi.refpinid = emptyref{}
		}
		return
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
		newnode, err = core.NewNode(Ctx, &core.BuildCfg{Online: true, Repo: newrepo})
	}
	var newapi iface.CoreAPI
	if err == nil {
		closeFunc = newnode.Close
		newapi, err = coreapi.NewCoreAPI(newnode)
	}
	if err == nil {
		nodeApi = newapi
		buildCount = 1
		mapi = &MfsAPI{
			lock:      &sync.RWMutex{},
			mroot:     nil,
			newcid:    nil,
			pinclient: pinningservice.NewClient(purl, ptoken),
			pinlock:   &sync.Mutex{},
			pinning:   []chan<- error{},
			queued:    []chan<- error{},
			refcid:    rcid,
			refpinid:  rpinid,
		}
		mapi.ctx, mapi.cancel = context.WithCancel(Ctx)
		if mapi.refcid == nil {
			mapi.refcid = emptyref{}
		}
		if mapi.refpinid == nil {
			mapi.refpinid = emptyref{}
		}
	}
	return
}
func (mapi *MfsAPI) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	mapi.cancel()
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
	ctxerr := mapi.ctx.Err()
	mapi.pinlock.Lock()
	mapi.queued = append(mapi.pinning, mapi.queued...)
	mapi.pinning = []chan<- error{}
	for _, v := range mapi.queued {
		v <- ctxerr
		close(v)
	}
	mapi.queued = []chan<- error{}
	if mapi.mroot != nil {
		err = mapi.mroot.FlushMemFree(mapi.ctx)
		mapi.mroot = nil
	}
	return
}
func (mapi *MfsAPI) runPin() {
	defer recover()
	queuelen := 0
	waittime := 0
	pinid := mapi.refpinid.Get()
	for {
		select {
		case <-mapi.ctx.Done():
			return
		case <-time.After(time.Second * 2):
			func() {
				mapi.pinlock.Lock()
				defer mapi.pinlock.Unlock()
				if len(mapi.pinning) > 0 {
					if pinstatus, err := mapi.pinclient.GetStatusByID(mapi.ctx, pinid); err == nil {
						if info, err := peer.AddrInfosFromP2pAddrs(pinstatus.GetDelegates()...); err == nil {
							for _, a := range info {
								go nodeApi.Swarm().Connect(Ctx, a)
							}
						}
						if pinstatus.GetStatus() == pinningservice.StatusPinned {
							for _, v := range mapi.pinning {
								v <- nil
								close(v)
							}
							mapi.pinning = []chan<- error{}
							mapi.refcid.Set(pinstatus.GetPin().GetCid().String())
						} else if pinstatus.GetStatus() == pinningservice.StatusFailed {
							for _, v := range mapi.pinning {
								v <- fmt.Errorf("StatusFailed")
								close(v)
							}
							mapi.pinning = []chan<- error{}
						}
					}
				}
				if mapi.newcid != nil {
					if queuelen == len(mapi.queued) || waittime > 10 {
						nodeApi.Swarm().ListenAddrs(mapi.ctx)
						if pinstatus, err := mapi.pinclient.Replace(
							mapi.ctx, pinid, *mapi.newcid, pinningservice.PinOpts.WithOrigins(),
						); err == nil {
							if info, err := peer.AddrInfosFromP2pAddrs(pinstatus.GetDelegates()...); err == nil {
								for _, a := range info {
									go nodeApi.Swarm().Connect(Ctx, a)
								}
							}
							mapi.pinning = append(mapi.pinning, mapi.queued...)
							pinid = pinstatus.GetRequestId()
							mapi.refpinid.Set(pinid)
						} else {
							for _, v := range mapi.queued {
								v <- err
								close(v)
							}
						}
						mapi.newcid = nil
						mapi.queued = []chan<- error{}
						waittime = 0
					}
					queuelen = len(mapi.queued)
					waittime++
				}
			}()
		}
	}
}
func (mapi *MfsAPI) waitPin(newcid cid.Cid) (ec <-chan error) {
	echan := make(chan error)
	if err := mapi.ctx.Err(); err != nil {
		go func() {
			echan <- err
			close(echan)
		}()
		return echan
	}
	mapi.pinlock.Lock()
	defer mapi.pinlock.Unlock()
	mapi.newcid = &newcid
	mapi.queued = append(mapi.queued, echan)
	return echan
}
func (mapi *MfsAPI) newRoot(force bool) (err error) {
	if err = mapi.ctx.Err(); err != nil {
		return
	}
	if !force && mapi.mroot != nil {
		return
	}
	mapi.lock.Lock()
	defer mapi.lock.Unlock()
	if !force && mapi.mroot != nil {
		return
	}
	rootcid := mapi.refcid.Get()
	if pinstatus, err := mapi.pinclient.GetStatusByID(mapi.ctx, mapi.refpinid.Get()); err == nil {
		if info, err := peer.AddrInfosFromP2pAddrs(pinstatus.GetDelegates()...); err == nil {
			for _, a := range info {
				go nodeApi.Swarm().Connect(Ctx, a)
			}
		}
		if pinstatus.GetStatus() == pinningservice.StatusPinned {
			rootcid = pinstatus.GetPin().GetCid().String()
		}
	}
	var ldnode ipldformat.Node
	if err == nil {
		ldnode, err = nodeApi.ResolveNode(mapi.ctx, ifacepath.New(rootcid))
	}
	prnode := &merkledag.ProtoNode{}
	if err == nil {
		ok := true
		if prnode, ok = ldnode.(*merkledag.ProtoNode); !ok {
			err = fmt.Errorf(reflect.TypeOf(ldnode).String())
		}
	}
	mroot := &mfs.Root{}
	if err == nil {
		mroot, err = mfs.NewRoot(mapi.ctx, nodeApi.Dag(), prnode, nil)
	}
	if err == nil {
		ldnode, err = mroot.GetDirectory().GetNode()
	}
	if err == nil {
		if mapi.mroot != nil {
			mapi.mroot.FlushMemFree(mapi.ctx)
		}
		mapi.mroot = mroot
		mapi.refcid.Set(ldnode.Cid().String())
	}
	return
}
func (mapi *MfsAPI) List(pth string) (ol []NodeObj, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	if err = mapi.newRoot(false); err != nil {
		return
	}
	mapi.lock.RLock()
	defer mapi.lock.RUnlock()
	snode, err := mfs.Lookup(mapi.mroot, pth)
	dnode, ok := snode.(*mfs.Directory)
	if err == nil && !ok {
		err = fmt.Errorf(reflect.TypeOf(snode).String())
	}
	nl := []mfs.NodeListing{}
	if err == nil {
		nl, err = dnode.List(mapi.ctx)
	}
	if err == nil {
		ol = []NodeObj{}
		for _, n := range nl {
			ol = append(ol, NodeObj{
				Id:    n.Hash,
				Name:  n.Name,
				Size:  n.Size,
				Isdir: n.Type == int(mfs.TDir),
			})
		}
	}
	return ol, err
}
func (mapi *MfsAPI) Mkdir(pth string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	if err = mapi.newRoot(false); err != nil {
		return
	}
	newcid := cid.Cid{}
	defer func() {
		if err == nil {
			err = <-mapi.waitPin(newcid)
		}
	}()
	mapi.lock.RLock()
	defer mapi.lock.RUnlock()
	if err = mfs.Mkdir(mapi.mroot, pth, mfs.MkdirOpts{}); err == nil {
		err = mapi.mroot.Flush()
	}
	if err == nil {
		var ldnode ipldformat.Node
		if ldnode, err = mapi.mroot.GetDirectory().GetNode(); err == nil {
			newcid = ldnode.Cid()
		}
	}
	return
}
func (mapi *MfsAPI) Mv(src, dst string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	if err = mapi.newRoot(false); err != nil {
		return
	}
	newcid := cid.Cid{}
	defer func() {
		if err == nil {
			err = <-mapi.waitPin(newcid)
		}
	}()
	mapi.lock.RLock()
	defer mapi.lock.RUnlock()
	if err = mfs.Mv(mapi.mroot, src, dst); err == nil {
		err = mapi.mroot.Flush()
	}
	if err == nil {
		var ldnode ipldformat.Node
		if ldnode, err = mapi.mroot.GetDirectory().GetNode(); err == nil {
			newcid = ldnode.Cid()
		}
	}
	return
}
func (mapi *MfsAPI) Put(pth, nodecid string, rc io.ReadCloser) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	if err = mapi.newRoot(false); err != nil {
		return
	}
	newcid := cid.Cid{}
	defer func() {
		if err == nil {
			err = <-mapi.waitPin(newcid)
		}
	}()
	mapi.lock.RLock()
	defer mapi.lock.RUnlock()
	var rsnode ipldformat.Node
	if rsnode, err = nodeApi.ResolveNode(mapi.ctx, ifacepath.New(nodecid)); err != nil {
		var rspath ifacepath.Resolved
		if rspath, err = nodeApi.Unixfs().Add(mapi.ctx, files.NewReaderFile(rc)); err == nil {
			rsnode, err = nodeApi.ResolveNode(mapi.ctx, rspath)
		}
	}
	if err == nil {
		err = mfs.PutNode(mapi.mroot, pth, rsnode)
	}
	if err == nil {
		err = mapi.mroot.Flush()
	}
	if err == nil {
		var ldnode ipldformat.Node
		if ldnode, err = mapi.mroot.GetDirectory().GetNode(); err == nil {
			newcid = ldnode.Cid()
		}
	}
	return
}
func (mapi *MfsAPI) Unlink(pth, fname string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	if err = mapi.newRoot(false); err != nil {
		return
	}
	newcid := cid.Cid{}
	defer func() {
		if err == nil {
			err = <-mapi.waitPin(newcid)
		}
	}()
	mapi.lock.RLock()
	defer mapi.lock.RUnlock()
	snode, err := mfs.Lookup(mapi.mroot, pth)
	dnode, ok := snode.(*mfs.Directory)
	if err == nil && !ok {
		err = fmt.Errorf(reflect.TypeOf(snode).String())
	}
	if err == nil {
		err = dnode.Unlink(fname)
	}
	if err == nil {
		err = mapi.mroot.Flush()
	}
	if err == nil {
		var ldnode ipldformat.Node
		if ldnode, err = mapi.mroot.GetDirectory().GetNode(); err == nil {
			newcid = ldnode.Cid()
		}
	}
	return
}
