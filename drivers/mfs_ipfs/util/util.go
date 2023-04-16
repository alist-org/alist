package util

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
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
	"github.com/multiformats/go-multiaddr"
)

type MfsAPI struct {
	cancel    context.CancelFunc
	ctx       context.Context
	lock      *sync.RWMutex
	log       *log.Logger
	mroot     *mfs.Root
	pinclient *pinningservice.Client
	qchan     chan chan<- error
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

func NewMfs(purl, ptoken string, rcid, rpinid refresher, fslog *log.Logger) (mapi *MfsAPI, err error) {
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
	if rcid == nil {
		rcid = emptyref{}
	}
	if rpinid == nil {
		rpinid = emptyref{}
	}
	if fslog == nil {
		fslog = log.New(io.Discard, "", log.LstdFlags)
	}

	// singleton
	if buildCount >= 0 {
		buildCount++
		mapi = &MfsAPI{
			lock:      &sync.RWMutex{},
			log:       fslog,
			mroot:     nil,
			pinclient: pinningservice.NewClient(purl, ptoken),
			qchan:     make(chan chan<- error),
			refcid:    rcid,
			refpinid:  rpinid,
		}
		mapi.ctx, mapi.cancel = context.WithCancel(Ctx)
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

	// setup repo
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

	// setup node
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
			log:       fslog,
			mroot:     nil,
			pinclient: pinningservice.NewClient(purl, ptoken),
			qchan:     make(chan chan<- error),
			refcid:    rcid,
			refpinid:  rpinid,
		}
		mapi.ctx, mapi.cancel = context.WithCancel(Ctx)
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
	if mapi.mroot != nil {
		err = mapi.mroot.FlushMemFree(mapi.ctx)
		mapi.mroot = nil
	}
	return
}

func (mapi *MfsAPI) runPin() {
	defer recover()
	var newcid *cid.Cid = nil
	pinid := mapi.refpinid.Get()
	pinning := []chan<- error{}
	queued := []chan<- error{}
	queuelen := 0
	waitnum := 0
	waittime := time.NewTimer(0)
	defer func() {
		waittime.Stop()
		err := mapi.ctx.Err()
		if err == nil {
			err = fmt.Errorf("Failed")
		}
		for _, v := range pinning {
			v <- err
		}
		pinning = []chan<- error{}
		for _, v := range queued {
			v <- err
		}
		queued = []chan<- error{}
		close(mapi.qchan)
		for v := range mapi.qchan {
			v <- err
		}
	}()
	for {
		select {
		case <-mapi.ctx.Done():
			return
		case <-waittime.C:
			if mapi.ctx.Err() != nil {
				return
			}
			func() {
				mapi.lock.RLock()
				defer mapi.lock.RUnlock()
				// get status
				if len(pinning) > 0 {
					if pinstatus, err := mapi.pinclient.GetStatusByID(mapi.ctx, pinid); err == nil {
						if info, err := peer.AddrInfosFromP2pAddrs(pinstatus.GetDelegates()...); err == nil {
							for _, a := range info {
								go nodeApi.Swarm().Connect(Ctx, a)
							}
						}
						if pinstatus.GetStatus() == pinningservice.StatusPinned {
							for _, v := range pinning {
								v <- nil
							}
							pinning = []chan<- error{}
							mapi.refcid.Set(pinstatus.GetPin().GetCid().String())
						} else if pinstatus.GetStatus() == pinningservice.StatusFailed {
							for _, v := range pinning {
								v <- fmt.Errorf("StatusFailed")
							}
							pinning = []chan<- error{}
						}
						mapi.log.Println(pinstatus.GetPin().GetCid(), pinstatus.GetStatus())
					} else {
						mapi.log.Println(err)
					}
				}
				if len(queued) > 0 {
					if waitnum++; len(queued) == queuelen || waitnum > 10 {
						if ldnode, err := mapi.mroot.GetDirectory().GetNode(); err == nil {
							pincid := ldnode.Cid()
							newcid = &pincid
							waitnum = 0
						}
					}
				}
				// replace pin
				if newcid != nil {
					oriaddr, _ := nodeApi.Swarm().ListenAddrs(mapi.ctx)
					mafilter := multiaddr.NewFilters()
					mafilter.DefaultAction = multiaddr.ActionDeny
					for _, v := range []string{"127.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12", "192.168.0.0/16", "::1/128"} {
						if _, ipnet, err := net.ParseCIDR(v); err == nil {
							mafilter.AddFilter(*ipnet, multiaddr.ActionAccept)
						}
					}
					oriaddr = multiaddr.FilterAddrs(oriaddr, mafilter.AddrBlocked)
					if pinstatus, err := mapi.pinclient.Replace(
						mapi.ctx, pinid, *newcid, pinningservice.PinOpts.WithOrigins(oriaddr...),
					); err == nil {
						if info, err := peer.AddrInfosFromP2pAddrs(pinstatus.GetDelegates()...); err == nil {
							for _, a := range info {
								go nodeApi.Swarm().Connect(Ctx, a)
							}
						}
						pinning = append(pinning, queued...)
						pinid = pinstatus.GetRequestId()
						mapi.refpinid.Set(pinid)
					} else {
						for _, v := range queued {
							v <- err
						}
					}
					queued = []chan<- error{}
					newcid = nil
				}
				queuelen = len(queued)
			}()
			waittime.Reset(time.Second * 2)
		case newpin := <-mapi.qchan:
			queued = append(queued, newpin)
		}
	}
}
func (mapi *MfsAPI) waitPin(pe *error) {
	if err := mapi.ctx.Err(); err != nil && *pe == nil {
		*pe = err
	}
	if *pe == nil {
		echan := make(chan error)
		mapi.qchan <- echan
		*pe = <-echan
	}
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
	defer mapi.waitPin(&err)
	mapi.lock.RLock()
	defer mapi.lock.RUnlock()
	err = mfs.Mkdir(mapi.mroot, pth, mfs.MkdirOpts{})
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
	defer mapi.waitPin(&err)
	mapi.lock.RLock()
	defer mapi.lock.RUnlock()
	err = mfs.Mv(mapi.mroot, src, dst)
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
	defer mapi.waitPin(&err)
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
	defer mapi.waitPin(&err)
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
	return
}
