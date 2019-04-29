# 1. [DC-06] Crash by Negative Height

## 1.1. Tag

`negative height`;

`crash`;

`replay`

## 1.2. Vulnerability description

Crash by Negative Height

Block height is defined in consensus, node, communication modules. Blockchain height should not be set as a negative value. Though it allows to input of negative value, the value should be checked carefully.

## 1.3. Vulnerability analysis

File location : `consensus/replay.go`

```golang
    blockHeight := int64(res.LastBlockHeight)

    // codes ...

	_, err = h.ReplayBlocks(appHash, blockHeight, proxyApp)
	if err != nil {
		return errors.New(cmn.Fmt("Error on replay: %v", err))
	}
```

## 1.4. Reproduce

### 1.4.1. Use go test script to test

Here we use the testing script `nega_height_test.go` to test. You can create a file anywhere named `nega_height_test.go` and copy the content below, run it in the terminal.

Before writing test file, let's first to add some auxiliary code to the real code to help us test.

```go
// XuanMao: mock for test, add function MockRespInfo in /consensus/replay.go
func MockRespInfo(height int64) *abci.ResponseInfo {
    return &abci.ResponseInfo{
        Data: "testdata",
        Version: "v0.1",
        LastBlockHeight: height,
    }
}

// around line 230
func (h *Handshaker) Handshake(proxyApp proxy.AppConns) error {

    // Handshake is done via ABCI Info on the query conn.
    res, err := proxyApp.Query().InfoSync(proxy.RequestInfo)
    if err != nil {
        return fmt.Errorf("Error calling Info: %v", err)
    }
    res = MockRespInfo(-1) // XuanMao: for test
    blockHeight := int64(res.LastBlockHeight)

```

Now we can write our testing file.

```golang
// XuanMao: bug test
// nega_height_test.go
package consensus

import (
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "path"
    "runtime"
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/XuanMaoSecLab/DolphinChain/abci/example/kvstore"
    abci "github.com/XuanMaoSecLab/DolphinChain/abci/types"
    cfg "github.com/XuanMaoSecLab/DolphinChain/config"
    "github.com/XuanMaoSecLab/DolphinChain/crypto"
    dbm "github.com/XuanMaoSecLab/DolphinChain/libs/db"
    "github.com/XuanMaoSecLab/DolphinChain/libs/log"
    "github.com/XuanMaoSecLab/DolphinChain/privval"
    "github.com/XuanMaoSecLab/DolphinChain/proxy"
    sm "github.com/XuanMaoSecLab/DolphinChain/state"
    "github.com/XuanMaoSecLab/DolphinChain/types"
    "github.com/XuanMaoSecLab/DolphinChain/version"
    . "github.com/XuanMaoSecLab/DolphinChain/consensus"
)

type crashingWAL struct {
    next         WAL
    panicCh      chan error
    heightToStop int64

    msgIndex                int // current message index
    lastPanickedForMsgIndex int // last message for which we panicked
}

var _ WAL = &crashingWAL{}

// WALWriteError indicates a WAL crash.
type WALWriteError struct {
    msg string
}

func (e WALWriteError) Error() string {
    return e.msg
}

// ReachedHeightToStopError indicates we've reached the required consensus
// height and may exit.
type ReachedHeightToStopError struct {
    height int64
}

func (e ReachedHeightToStopError) Error() string {
    return fmt.Sprintf("reached height to stop %d", e.height)
}

// Write simulate WAL's crashing by sending an error to the panicCh and then
// exiting the cs.receiveRoutine.
func (w *crashingWAL) Write(m WALMessage) {
    if endMsg, ok := m.(EndHeightMessage); ok {
        if endMsg.Height == w.heightToStop {
            w.panicCh <- ReachedHeightToStopError{endMsg.Height}
            runtime.Goexit()
        } else {
            w.next.Write(m)
        }
        return
    }

    if w.msgIndex > w.lastPanickedForMsgIndex {
        w.lastPanickedForMsgIndex = w.msgIndex
        _, file, line, _ := runtime.Caller(1)
        w.panicCh <- WALWriteError{fmt.Sprintf("failed to write %T to WAL (fileline: %s:%d)", m, file, line)}
        runtime.Goexit()
    } else {
        w.msgIndex++
        w.next.Write(m)
    }
}

func (w *crashingWAL) WriteSync(m WALMessage) {
    w.Write(m)
}

func (w *crashingWAL) FlushAndSync() error { return w.next.FlushAndSync() }

func (w *crashingWAL) SearchForEndHeight(height int64, options *WALSearchOptions) (rd io.ReadCloser, found bool, err error) {
    return w.next.SearchForEndHeight(height, options)
}

func (w *crashingWAL) Start() error { return w.next.Start() }
func (w *crashingWAL) Stop() error  { return w.next.Stop() }
func (w *crashingWAL) Wait()        { w.next.Wait() }

//------------------------------------------------------------------------------------------
// Handshake Tests

const (
    NUM_BLOCKS = 6
)

var (
    mempool = sm.MockMempool{}
    evpool  = sm.MockEvidencePool{}
)

//---------------------------------------
// Test handshake/replay

// 0 - all synced up
// 1 - saved block but app and state are behind
// 2 - save block and committed but state is behind
var modes = []uint{0, 1, 2}

func tempWALWithData(data []byte) string {
    walFile, err := ioutil.TempFile("", "wal")
    if err != nil {
        panic(fmt.Errorf("failed to create temp WAL file: %v", err))
    }
    _, err = walFile.Write(data)
    if err != nil {
        panic(fmt.Errorf("failed to write to temp WAL file: %v", err))
    }
    if err := walFile.Close(); err != nil {
        panic(fmt.Errorf("failed to close temp WAL file: %v", err))
    }
    return walFile.Name()
}

// Make some blocks. Start a fresh app and apply nBlocks blocks. Then restart the app and sync it up with the remaining blocks

func applyBlock(stateDB dbm.DB, st sm.State, blk *types.Block, proxyApp proxy.AppConns) sm.State {
    testPartSize := types.BlockPartSizeBytes
    blockExec := sm.NewBlockExecutor(stateDB, log.TestingLogger(), proxyApp.Consensus(), mempool, evpool)

    blkID := types.BlockID{blk.Hash(), blk.MakePartSet(testPartSize).Header()}
    newState, err := blockExec.ApplyBlock(st, blkID, blk)
    if err != nil {
        panic(err)
    }
    return newState
}

func buildAppStateFromChain(proxyApp proxy.AppConns, stateDB dbm.DB,
    state sm.State, chain []*types.Block, nBlocks int, mode uint) {
    // start a new app without handshake, play nBlocks blocks
    if err := proxyApp.Start(); err != nil {
        panic(err)
    }
    defer proxyApp.Stop()

    validators := types.TM2PB.ValidatorUpdates(state.Validators)
    if _, err := proxyApp.Consensus().InitChainSync(abci.RequestInitChain{
        Validators: validators,
    }); err != nil {
        panic(err)
    }

    switch mode {
    case 0:
        for i := 0; i < nBlocks; i++ {
            block := chain[i]
            state = applyBlock(stateDB, state, block, proxyApp)
        }
    case 1, 2:
        for i := 0; i < nBlocks-1; i++ {
            block := chain[i]
            state = applyBlock(stateDB, state, block, proxyApp)
        }

        if mode == 2 {
            // update the kvstore height and apphash
            // as if we ran commit but not
            state = applyBlock(stateDB, state, chain[nBlocks-1], proxyApp)
        }
    }

}

func buildTMStateFromChain(config *cfg.Config, stateDB dbm.DB, state sm.State, chain []*types.Block, mode uint) sm.State {
    // run the whole chain against this client to build up the tendermint state
    clientCreator := proxy.NewLocalClientCreator(kvstore.NewPersistentKVStoreApplication(path.Join(config.DBDir(), "1")))
    proxyApp := proxy.NewAppConns(clientCreator)
    if err := proxyApp.Start(); err != nil {
        panic(err)
    }
    defer proxyApp.Stop()

    validators := types.TM2PB.ValidatorUpdates(state.Validators)
    if _, err := proxyApp.Consensus().InitChainSync(abci.RequestInitChain{
        Validators: validators,
    }); err != nil {
        panic(err)
    }

    switch mode {
    case 0:
        // sync right up
        for _, block := range chain {
            state = applyBlock(stateDB, state, block, proxyApp)
        }

    case 1, 2:
        // sync up to the penultimate as if we stored the block.
        // whether we commit or not depends on the appHash
        for _, block := range chain[:len(chain)-1] {
            state = applyBlock(stateDB, state, block, proxyApp)
        }

        // apply the final block to a state copy so we can
        // get the right next appHash but keep the state back
        applyBlock(stateDB, state, chain[len(chain)-1], proxyApp)
    }

    return state
}

//--------------------------
// utils for making blocks


// fresh state and mock store
func stateAndStore(config *cfg.Config, pubKey crypto.PubKey, appVersion version.Protocol) (dbm.DB, sm.State, *mockBlockStore) {
    stateDB := dbm.NewMemDB()
    state, _ := sm.MakeGenesisStateFromFile(config.GenesisFile())
    state.Version.Consensus.App = appVersion
    store := NewMockBlockStore(config, state.ConsensusParams)
    return stateDB, state, store
}

//----------------------------------
// mock block store

type mockBlockStore struct {
    config  *cfg.Config
    params  types.ConsensusParams
    chain   []*types.Block
    commits []*types.Commit
}

// TODO: NewBlockStore(db.NewMemDB) ...
func NewMockBlockStore(config *cfg.Config, params types.ConsensusParams) *mockBlockStore {
    return &mockBlockStore{config, params, nil, nil}
}

func (bs *mockBlockStore) Height() int64                       { return int64(len(bs.chain)) }
func (bs *mockBlockStore) LoadBlock(height int64) *types.Block { return bs.chain[height-1] }
func (bs *mockBlockStore) LoadBlockMeta(height int64) *types.BlockMeta {
    block := bs.chain[height-1]
    return &types.BlockMeta{
        BlockID: types.BlockID{block.Hash(), block.MakePartSet(types.BlockPartSizeBytes).Header()},
        Header:  block.Header,
    }
}
func (bs *mockBlockStore) LoadBlockPart(height int64, index int) *types.Part { return nil }
func (bs *mockBlockStore) SaveBlock(block *types.Block, blockParts *types.PartSet, seenCommit *types.Commit) {
}
func (bs *mockBlockStore) LoadBlockCommit(height int64) *types.Commit {
    return bs.commits[height-1]
}
func (bs *mockBlockStore) LoadSeenCommit(height int64) *types.Commit {
    return bs.commits[height-1]
}

//----------------------------------------

func TestInitChainUpdateValidators(t *testing.T) {
    val, _ := types.RandValidator(true, 10)
    vals := types.NewValidatorSet([]*types.Validator{val})
    app := &initChainApp{vals: types.TM2PB.ValidatorUpdates(vals)}
    clientCreator := proxy.NewLocalClientCreator(app)
    config := cfg.ResetTestRoot("proxy_test")
    defer os.RemoveAll(config.RootDir)
    privVal := privval.LoadFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile())
    stateDB, state, store := stateAndStore(config, privVal.GetPubKey(), 0x0)

    oldValAddr := state.Validators.Validators[0].Address

    // now start the app using the handshake - it should sync
    genDoc, _ := sm.MakeGenesisDocFromFile(config.GenesisFile())
    handshaker := NewHandshaker(stateDB, state, store, genDoc)
    proxyApp := proxy.NewAppConns(clientCreator)
    if err := proxyApp.Start(); err != nil {
        t.Fatalf("Error starting proxy app connections: %v", err)
    }
    defer proxyApp.Stop()
    if err := handshaker.Handshake(proxyApp); err != nil {
        t.Fatalf("Error on abci handshake: %v", err)
    }

    // reload the state, check the validator set was updated
    state = sm.LoadState(stateDB)

    newValAddr := state.Validators.Validators[0].Address
    expectValAddr := val.Address
    assert.NotNil(t, handshaker, oldValAddr, newValAddr, expectValAddr)
}

// returns the vals on InitChain
type initChainApp struct {
    abci.BaseApplication
    vals []abci.ValidatorUpdate
}

func (ica *initChainApp) InitChain(req abci.RequestInitChain) abci.ResponseInitChain {
    return abci.ResponseInitChain{
        Validators: ica.vals,
    }
}


```

### 1.4.2. Testing result

```sh
=== RUN   TestInitChainUpdateValidators
--- FAIL: TestInitChainUpdateValidators (0.01s)
    replay_test.go:310: Error on abci handshake: Got a negative last block height (-1) from the app
FAIL
exit status 1
FAIL    github.com/XuanMaoSecLab/DolphinChain/reproduce/nega_height_check    0.245s

```

## 1.5. Fix

Reference of this vulnerability: [Fix](https://github.com/tendermint/tendermint/commit/89cbcceac4d7359a4d0b38bedd137654279a006d)

### 1.5.1. Fix method

Add the fixed code in the function `@Handshaker` of `consensus/replay.go` to check the blockheight.

```golang
...
    blockHeight := int64(res.LastBlockHeight)
    if blockHeight < 0 {
        return fmt.Errorf("Got a negative last block height (%d) from the app", blockHeight)
    }
    appHash := res.LastBlockAppHash
```

## 1.6. Reference

Vulnerable code from [No check](https://github.com/tendermint/tendermint/commit/89cbcceac4d7359a4d0b38bedd137654279a006d)

You can check related issue [here](https://github.com/tendermint/tendermint/issues/911).
