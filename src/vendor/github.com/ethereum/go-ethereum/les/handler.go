// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package les implements the Light Ethereum Subprotocol.
package les

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/pow"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

const (
	softResponseLimit = 2 * 1024 * 1024 // Target maximum size of returned blocks, headers or node data.
	estHeaderRlpSize  = 500             // Approximate size of an RLP encoded block header

	ethVersion = 63 // equivalent eth version for the downloader

	MaxHeaderFetch  = 192 // Amount of block headers to be fetched per retrieval request
	MaxBodyFetch    = 32  // Amount of block bodies to be fetched per retrieval request
	MaxReceiptFetch = 128 // Amount of transaction receipts to allow fetching per request
	MaxCodeFetch    = 64  // Amount of contract codes to allow fetching per request
	MaxProofsFetch  = 64  // Amount of merkle proofs to be fetched per retrieval request
)

// errIncompatibleConfig is returned if the requested protocols and configs are
// not compatible (low protocol version restrictions and high requirements).
var errIncompatibleConfig = errors.New("incompatible configuration")

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

type hashFetcherFn func(common.Hash) error

type BlockChain interface {
	HasHeader(hash common.Hash) bool
	GetHeader(hash common.Hash) *types.Header
	CurrentHeader() *types.Header
	GetTd(hash common.Hash) *big.Int
	InsertHeaderChain(chain []*types.Header, checkFreq int) (int, error)
	Rollback(chain []common.Hash)
	Status() (td *big.Int, currentBlock common.Hash, genesisBlock common.Hash)
	GetHeaderByNumber(number uint64) *types.Header
	GetBlockHashesFromHash(hash common.Hash, max uint64) []common.Hash
	LastBlockHash() common.Hash
	Genesis() *types.Block
}

type txPool interface {
	// AddTransactions should add the given transactions to the pool.
	AddTransactions([]*types.Transaction)
}

type ProtocolManager struct {
	lightSync  bool
	txpool     txPool
	txrelay    *LesTxRelay
	networkId  int
	chainConfig *core.ChainConfig
	blockchain BlockChain
	chainDb    ethdb.Database
	odr        *LesOdr
	server     *LesServer

	downloader *downloader.Downloader
	//fetcher    *fetcher.Fetcher
	peers *peerSet

	SubProtocols []p2p.Protocol

	eventMux *event.TypeMux

	// channels for fetcher, syncer, txsyncLoop
	newPeerCh   chan *peer
	quitSync    chan struct{}
	noMorePeers chan struct{}

	// wait group is used for graceful shutdowns during downloading
	// and processing
	wg sync.WaitGroup
}

// NewProtocolManager returns a new ethereum sub protocol manager. The Ethereum sub protocol manages peers capable
// with the ethereum network.
func NewProtocolManager(chainConfig *core.ChainConfig, lightSync bool, networkId int, mux *event.TypeMux, pow pow.PoW, blockchain BlockChain, txpool txPool, chainDb ethdb.Database, odr *LesOdr, txrelay *LesTxRelay) (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		lightSync:   lightSync,
		eventMux:    mux,
		blockchain:  blockchain,
		chainConfig: chainConfig,
		chainDb:     chainDb,
		networkId:   networkId,
		txpool:      txpool,
		txrelay:     txrelay,
		odr:         odr,
		peers:       newPeerSet(),
		newPeerCh:   make(chan *peer, 1),
		quitSync:    make(chan struct{}),
		noMorePeers: make(chan struct{}),
	}
	// Initiate a sub-protocol for every implemented version we can handle
	manager.SubProtocols = make([]p2p.Protocol, 0, len(ProtocolVersions))
	for i, version := range ProtocolVersions {
		// Compatible, initialize the sub-protocol
		version := version // Closure for the run
		manager.SubProtocols = append(manager.SubProtocols, p2p.Protocol{
			Name:    "les",
			Version: version,
			Length:  ProtocolLengths[i],
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				peer := manager.newPeer(int(version), networkId, p, rw)
				select {
				case manager.newPeerCh <- peer:
					manager.wg.Add(1)
					defer manager.wg.Done()
					return manager.handle(peer)
				case <-manager.quitSync:
					return p2p.DiscQuitting
				}
			},
			NodeInfo: func() interface{} {
				return manager.NodeInfo()
			},
			PeerInfo: func(id discover.NodeID) interface{} {
				if p := manager.peers.Peer(fmt.Sprintf("%x", id[:8])); p != nil {
					return p.Info()
				}
				return nil
			},
		})
	}
	if len(manager.SubProtocols) == 0 {
		return nil, errIncompatibleConfig
	}

	if lightSync {
		glog.V(logger.Debug).Infof("LES: create downloader")
		manager.downloader = downloader.New(downloader.LightSync, chainDb, manager.eventMux, blockchain.HasHeader, nil, blockchain.GetHeader,
			nil, blockchain.CurrentHeader, nil, nil, nil, blockchain.GetTd,
			blockchain.InsertHeaderChain, nil, nil, blockchain.Rollback, manager.removePeer)
	}

	/*validator := func(block *types.Block, parent *types.Block) error {
		return core.ValidateHeader(pow, block.Header(), parent.Header(), true, false)
	}
	heighter := func() uint64 {
		return chainman.LastBlockNumberU64()
	}
	manager.fetcher = fetcher.New(chainman.GetBlockNoOdr, validator, nil, heighter, chainman.InsertChain, manager.removePeer)
	*/
	return manager, nil
}

func (pm *ProtocolManager) removePeer(id string) {
	// Short circuit if the peer was already removed
	peer := pm.peers.Peer(id)
	if peer == nil {
		return
	}
	glog.V(logger.Debug).Infoln("Removing peer", id)

	// Unregister the peer from the downloader and Ethereum peer set
	glog.V(logger.Debug).Infof("LES: unregister peer %v", id)
	if pm.lightSync {
		pm.downloader.UnregisterPeer(id)
		pm.odr.UnregisterPeer(peer)
		if pm.txrelay != nil {
			pm.txrelay.removePeer(id)
		}
	}
	if err := pm.peers.Unregister(id); err != nil {
		glog.V(logger.Error).Infoln("Removal failed:", err)
	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
	}
}

func (pm *ProtocolManager) Start() {
	if pm.lightSync {
		// start sync handler
		go pm.syncer()
	}
}

func (pm *ProtocolManager) Stop() {
	// Showing a log message. During download / process this could actually
	// take between 5 to 10 seconds and therefor feedback is required.
	glog.V(logger.Info).Infoln("Stopping light ethereum protocol handler...")

	// Quit the sync loop.
	// After this send has completed, no new peers will be accepted.
	pm.noMorePeers <- struct{}{}

	close(pm.quitSync) // quits syncer, fetcher

	// Disconnect existing sessions.
	// This also closes the gate for any new registrations on the peer set.
	// sessions which are already established but not added to pm.peers yet
	// will exit when they try to register.
	pm.peers.Close()

	// Wait for any process action
	pm.wg.Wait()

	glog.V(logger.Info).Infoln("Light ethereum protocol handler stopped")
}

func (pm *ProtocolManager) newPeer(pv, nv int, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return newPeer(pv, nv, p, newMeteredMsgWriter(rw))
}

// handle is the callback invoked to manage the life cycle of a les peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p *peer) error {
	glog.V(logger.Debug).Infof("%v: peer connected [%s]", p, p.Name())

	// Execute the LES handshake
	td, head, genesis := pm.blockchain.Status()
	if err := p.Handshake(td, head, genesis, pm.server); err != nil {
		glog.V(logger.Debug).Infof("%v: handshake failed: %v", p, err)
		return err
	}
	if rw, ok := p.rw.(*meteredMsgReadWriter); ok {
		rw.Init(p.version)
	}
	// Register the peer locally
	glog.V(logger.Detail).Infof("%v: adding peer", p)
	if err := pm.peers.Register(p); err != nil {
		glog.V(logger.Error).Infof("%v: addition failed: %v", p, err)
		return err
	}
	defer func() {
		if pm.server != nil && pm.server.fcManager != nil && p.fcClient != nil {
			p.fcClient.Remove(pm.server.fcManager)
		}
		pm.removePeer(p.id)
	}()

	// Register the peer in the downloader. If the downloader considers it banned, we disconnect
	glog.V(logger.Debug).Infof("LES: register peer %v", p.id)
	if pm.lightSync {
		requestHeadersByHash := func(origin common.Hash, amount int, skip int, reverse bool) error {
			reqID := pm.odr.getNextReqID()
			cost := p.GetRequestCost(GetBlockHeadersMsg, amount)
			p.fcServer.SendRequest(reqID, cost)
			return p.RequestHeadersByHash(reqID, cost, origin, amount, skip, reverse)
		}
		requestHeadersByNumber := func(origin uint64, amount int, skip int, reverse bool) error {
			reqID := pm.odr.getNextReqID()
			cost := p.GetRequestCost(GetBlockHeadersMsg, amount)
			p.fcServer.SendRequest(reqID, cost)
			return p.RequestHeadersByNumber(reqID, cost, origin, amount, skip, reverse)
		}
		if err := pm.downloader.RegisterPeer(p.id, ethVersion, p.Head(),
			nil, nil, nil, requestHeadersByHash, requestHeadersByNumber,
			nil, nil, nil); err != nil {
			return err
		}
		pm.odr.RegisterPeer(p)
		if pm.txrelay != nil {
			pm.txrelay.addPeer(p)
		}
	}

	// main loop. handle incoming messages.
	for {
		if err := pm.handleMsg(p); err != nil {
			glog.V(logger.Debug).Infof("%v: message handling failed: %v", p, err)
			return err
		}
	}
	return nil
}

func getHeadersBatch(db ethdb.Database, hash common.Hash, number uint64, amount uint64, skip uint64, reverse bool) (res []rlp.RawValue) {
	res = make([]rlp.RawValue, amount)
	step := int64(skip + 1)
	if reverse {
		step = -step
	}
	ptr := uint64(0)
	if hash != (common.Hash{}) {
		res[0] = core.GetHeaderRLP(db, hash)
		if res[0] == nil {
			return nil
		}
		header := new(types.Header)
		if err := rlp.Decode(bytes.NewReader(res[0]), header); err != nil {
			return nil
		}
		ptr = 1
		number = header.GetNumberU64()
	}

	tasks := make(chan uint64)
	go func() {
		stop := time.After(time.Millisecond * 300)
	loop:
		for i := ptr; i < amount; i++ {
			select {
			case tasks <- i:
			case <-stop:
				break loop
			}
		}
		close(tasks)
	}()

	pending := new(sync.WaitGroup)
	for i := 0; i < 4; i++ {
		pending.Add(1)
		go func(id int) {
			for index := range tasks {
				num := int64(number) + step*int64(index)
				if num >= 0 {
					hash := core.GetCanonicalHash(db, uint64(num))
					if hash != (common.Hash{}) {
						res[index] = core.GetHeaderRLP(db, hash)
					}
				}
			}
			pending.Done()
		}(i)
	}
	pending.Wait()

	for i := 0; i < len(res); i++ {
		if res[i] == nil {
			res = res[:i]
			return
		}
	}
	return
}

var reqList = []uint64{GetBlockHeadersMsg, GetBlockBodiesMsg, GetCodeMsg, GetReceiptsMsg, GetProofsMsg, SendTxMsg}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (pm *ProtocolManager) handleMsg(p *peer) error {
	// Read the next message from the remote peer, and ensure it's fully consumed
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}

	var costs *requestCosts
	var reqCnt, maxReqs int

	if rc, ok := p.fcCosts[msg.Code]; ok { // check if msg is a supported request type
		costs = rc
		if p.fcClient == nil {
			return errResp(ErrRequestRejected, "")
		}
		bv, ok := p.fcClient.AcceptRequest()
		if !ok || bv < costs.baseCost {
			return errResp(ErrRequestRejected, "")
		}
		d := bv - costs.baseCost
		if d/10000 < costs.reqCost {
			maxReqs = int(d / costs.reqCost)
		} else {
			maxReqs = 10000
		}
	}

	if msg.Size > ProtocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	defer msg.Discard()

	var deliverMsg *Msg

	// Handle the message depending on its contents
	switch msg.Code {
	case StatusMsg:
		glog.V(logger.Debug).Infof("LES: received StatusMsg from peer %v", p.id)
		// Status messages should never arrive after the handshake
		return errResp(ErrExtraStatusMsg, "uncontrolled status message")

	// Block header query, collect the requested headers and reply
	case GetBlockHeadersMsg:
		glog.V(logger.Debug).Infof("LES: received GetBlockHeadersMsg from peer %v", p.id)
		// Decode the complex header query
		var query struct {
			ReqID uint64
			getBlockHeadersData
		}
		if err := msg.Decode(&query); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		amount := query.Amount
		if amount > uint64(maxReqs) {
			return errResp(ErrRequestRejected, "")
		}
		if amount > MaxHeaderFetch {
			amount = MaxHeaderFetch
		}
		if amount*estHeaderRlpSize > softResponseLimit {
			amount = softResponseLimit / estHeaderRlpSize
		}

		headers := getHeadersBatch(pm.chainDb, query.Origin.Hash, query.Origin.Number, amount, query.Skip, query.Reverse)
		bv, rcost := p.fcClient.RequestProcessed(costs.baseCost + amount*costs.reqCost)
		pm.server.fcCostStats.update(msg.Code, amount, rcost)
		return p.SendBlockHeaders(query.ReqID, bv, headers)

	case BlockHeadersMsg:
		if pm.downloader == nil {
			return errResp(ErrUnexpectedResponse, "")
		}

		glog.V(logger.Debug).Infof("LES: received BlockHeadersMsg from peer %v", p.id)
		// A batch of headers arrived to one of our previous requests
		var resp struct {
			ReqID, BV uint64
			Headers   []*types.Header
		}
		if err := msg.Decode(&resp); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		p.fcServer.GotReply(resp.ReqID, resp.BV)
		err := pm.downloader.DeliverHeaders(p.id, resp.Headers)
		if err != nil {
			glog.V(logger.Debug).Infoln(err)
		}

	case GetBlockBodiesMsg:
		glog.V(logger.Debug).Infof("LES: received GetBlockBodiesMsg from peer %v", p.id)
		// Decode the retrieval message
		var req struct {
			ReqID  uint64
			Hashes []common.Hash
		}
		if err := msg.Decode(&req); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Gather blocks until the fetch or network limits is reached
		var (
			bytes  int
			bodies []rlp.RawValue
		)
		reqCnt = len(req.Hashes)
		if reqCnt > maxReqs {
			return errResp(ErrRequestRejected, "")
		}
		for _, hash := range req.Hashes {
			if bytes >= softResponseLimit || len(bodies) >= MaxBodyFetch {
				break
			}
			// Retrieve the requested block body, stopping if enough was found
			if data := core.GetBodyRLP(pm.chainDb, hash); len(data) != 0 {
				bodies = append(bodies, data)
				bytes += len(data)
			}
		}
		bv, rcost := p.fcClient.RequestProcessed(costs.baseCost + uint64(reqCnt)*costs.reqCost)
		pm.server.fcCostStats.update(msg.Code, uint64(reqCnt), rcost)
		return p.SendBlockBodiesRLP(req.ReqID, bv, bodies)

	case BlockBodiesMsg:
		if pm.odr == nil {
			return errResp(ErrUnexpectedResponse, "")
		}

		glog.V(logger.Debug).Infof("LES: received BlockBodiesMsg from peer %v", p.id)
		// A batch of block bodies arrived to one of our previous requests
		var resp struct {
			ReqID, BV uint64
			Data      []*types.Body
		}
		if err := msg.Decode(&resp); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		p.fcServer.GotReply(resp.ReqID, resp.BV)
		deliverMsg = &Msg{
			MsgType: MsgBlockBodies,
			ReqID:   resp.ReqID,
			Obj:     resp.Data,
		}

	case GetCodeMsg:
		glog.V(logger.Debug).Infof("LES: received GetCodeMsg from peer %v", p.id)
		// Decode the retrieval message
		var req struct {
			ReqID uint64
			Reqs  []CodeReq
		}
		if err := msg.Decode(&req); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Gather state data until the fetch or network limits is reached
		var (
			bytes int
			data  [][]byte
		)
		reqCnt = len(req.Reqs)
		if reqCnt > maxReqs {
			return errResp(ErrRequestRejected, "")
		}
		for _, req := range req.Reqs {
			if len(data) >= MaxCodeFetch {
				break
			}
			// Retrieve the requested state entry, stopping if enough was found
			if header := core.GetHeader(pm.chainDb, req.BHash); header != nil {
				if trie, _ := trie.New(header.Root, pm.chainDb); trie != nil {
					sdata := trie.Get(req.AccKey)
					if so, err := state.DecodeObject(common.Address{}, pm.chainDb, sdata); err == nil {
						entry := so.Code()
						if bytes+len(entry) >= softResponseLimit {
							break
						}
						data = append(data, entry)
						bytes += len(entry)
					}
				}
			}
		}
		bv, rcost := p.fcClient.RequestProcessed(costs.baseCost + uint64(reqCnt)*costs.reqCost)
		pm.server.fcCostStats.update(msg.Code, uint64(reqCnt), rcost)
		return p.SendCode(req.ReqID, bv, data)

	case CodeMsg:
		if pm.odr == nil {
			return errResp(ErrUnexpectedResponse, "")
		}

		glog.V(logger.Debug).Infof("LES: received CodeMsg from peer %v", p.id)
		// A batch of node state data arrived to one of our previous requests
		var resp struct {
			ReqID, BV uint64
			Data      [][]byte
		}
		if err := msg.Decode(&resp); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		p.fcServer.GotReply(resp.ReqID, resp.BV)
		deliverMsg = &Msg{
			MsgType: MsgCode,
			ReqID:   resp.ReqID,
			Obj:     resp.Data,
		}

	case GetReceiptsMsg:
		glog.V(logger.Debug).Infof("LES: received GetReceiptsMsg from peer %v", p.id)
		// Decode the retrieval message
		var req struct {
			ReqID  uint64
			Hashes []common.Hash
		}
		if err := msg.Decode(&req); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Gather state data until the fetch or network limits is reached
		var (
			bytes    int
			receipts []rlp.RawValue
		)
		reqCnt = len(req.Hashes)
		if reqCnt > maxReqs {
			return errResp(ErrRequestRejected, "")
		}
		for _, hash := range req.Hashes {
			if bytes >= softResponseLimit || len(receipts) >= MaxReceiptFetch {
				break
			}
			// Retrieve the requested block's receipts, skipping if unknown to us
			results := core.GetBlockReceipts(pm.chainDb, hash)
			if results == nil {
				if header := pm.blockchain.GetHeader(hash); header == nil || header.ReceiptHash != types.EmptyRootHash {
					continue
				}
			}
			// If known, encode and queue for response packet
			if encoded, err := rlp.EncodeToBytes(results); err != nil {
				glog.V(logger.Error).Infof("failed to encode receipt: %v", err)
			} else {
				receipts = append(receipts, encoded)
				bytes += len(encoded)
			}
		}
		bv, rcost := p.fcClient.RequestProcessed(costs.baseCost + uint64(reqCnt)*costs.reqCost)
		pm.server.fcCostStats.update(msg.Code, uint64(reqCnt), rcost)
		return p.SendReceiptsRLP(req.ReqID, bv, receipts)

	case ReceiptsMsg:
		if pm.odr == nil {
			return errResp(ErrUnexpectedResponse, "")
		}

		glog.V(logger.Debug).Infof("LES: received ReceiptsMsg from peer %v", p.id)
		// A batch of receipts arrived to one of our previous requests
		var resp struct {
			ReqID, BV uint64
			Receipts  []types.Receipts
		}
		if err := msg.Decode(&resp); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		p.fcServer.GotReply(resp.ReqID, resp.BV)
		deliverMsg = &Msg{
			MsgType: MsgReceipts,
			ReqID:   resp.ReqID,
			Obj:     resp.Receipts,
		}

	case GetProofsMsg:
		glog.V(logger.Debug).Infof("LES: received GetProofsMsg from peer %v", p.id)
		// Decode the retrieval message
		var req struct {
			ReqID uint64
			Reqs  []ProofReq
		}
		if err := msg.Decode(&req); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Gather state data until the fetch or network limits is reached
		var (
			bytes  int
			proofs proofsData
		)
		reqCnt = len(req.Reqs)
		if reqCnt > maxReqs {
			return errResp(ErrRequestRejected, "")
		}
		for _, req := range req.Reqs {
			if bytes >= softResponseLimit || len(proofs) >= MaxProofsFetch {
				break
			}
			// Retrieve the requested state entry, stopping if enough was found
			if header := core.GetHeader(pm.chainDb, req.BHash); header != nil {
				if tr, _ := trie.New(header.Root, pm.chainDb); tr != nil {
					if len(req.AccKey) > 0 {
						data := tr.Get(req.AccKey)
						tr = nil
						if so, err := state.DecodeObject(common.Address{}, pm.chainDb, data); err == nil {
							tr, _ = trie.New(common.BytesToHash(so.Root()), pm.chainDb)
						}
					}
					if tr != nil {
						proof := tr.Prove(req.Key)
						proofs = append(proofs, proof)
						bytes += len(proof)
					}
				}
			}
		}
		bv, rcost := p.fcClient.RequestProcessed(costs.baseCost + uint64(reqCnt)*costs.reqCost)
		pm.server.fcCostStats.update(msg.Code, uint64(reqCnt), rcost)
		return p.SendProofs(req.ReqID, bv, proofs)

	case ProofsMsg:
		if pm.odr == nil {
			return errResp(ErrUnexpectedResponse, "")
		}

		glog.V(logger.Debug).Infof("LES: received ProofsMsg from peer %v", p.id)
		// A batch of merkle proofs arrived to one of our previous requests
		var resp struct {
			ReqID, BV uint64
			Data      [][]rlp.RawValue
		}
		if err := msg.Decode(&resp); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		p.fcServer.GotReply(resp.ReqID, resp.BV)
		deliverMsg = &Msg{
			MsgType: MsgProofs,
			ReqID:   resp.ReqID,
			Obj:     resp.Data,
		}

	case SendTxMsg:
		if pm.txpool == nil {
			return errResp(ErrUnexpectedResponse, "")
		}
		// Transactions arrived, parse all of them and deliver to the pool
		var txs []*types.Transaction
		if err := msg.Decode(&txs); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		reqCnt = len(txs)
		if reqCnt > maxReqs {
			return errResp(ErrRequestRejected, "")
		}
		pm.txpool.AddTransactions(txs)
		_, rcost := p.fcClient.RequestProcessed(costs.baseCost + uint64(reqCnt)*costs.reqCost)
		pm.server.fcCostStats.update(msg.Code, uint64(reqCnt), rcost)

	default:
		glog.V(logger.Debug).Infof("LES: received unknown message with code %d from peer %v", msg.Code, p.id)
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}

	if deliverMsg != nil {
		return pm.odr.Deliver(p, deliverMsg)
	}

	return nil
}

// NodeInfo retrieves some protocol metadata about the running host node.
func (self *ProtocolManager) NodeInfo() *eth.EthNodeInfo {
	return &eth.EthNodeInfo{
		Network:    self.networkId,
		Difficulty: self.blockchain.GetTd(self.blockchain.LastBlockHash()),
		Genesis:    self.blockchain.Genesis().Hash(),
		Head:       self.blockchain.LastBlockHash(),
	}
}