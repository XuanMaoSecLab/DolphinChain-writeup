# [DC-06] 高度为负值导致崩溃

## 漏洞标签

`negative height`;

`crash`;

`replay`

## 漏洞描述

高度为负值导致崩溃(Crash by Negative Height)

区块高度在共识模块，节点模块，通信模块都有定义。区块高度不应为负值。即使可能输入为负值，需要进行验证

## 漏洞分析

文件位置 : `consensus/replay.go`

```golang
    blockHeight := int64(res.LastBlockHeight)

    // codes ...

	_, err = h.ReplayBlocks(appHash, blockHeight, proxyApp)
	if err != nil {
		return errors.New(cmn.Fmt("Error on replay: %v", err))
	}
```

## 复现或测试步骤

此处使用test脚本测试

### 使用 go test 脚本测试

为了方便测试，修改部分源码文件

```go
// XuanMao: mock for test, add function MockRespInfo in /consensus/replay.go
func MockRespInfo(height int64) *abci.ResponseInfo {
    return &abci.ResponseInfo{
        Data: "testdata",
        Version: "v0.1",
        LastBlockHeight: height,
    }
}

// TODO: retry the handshake/replay if it fails ?
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

```golang
// XuanMao: bug test
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

```

运行结果

```sh
=== RUN   TestInitChainUpdateValidators
--- FAIL: TestInitChainUpdateValidators (0.01s)
    replay_test.go:310: Error on abci handshake: Got a negative last block height (-1) from the app
FAIL
exit status 1
FAIL	github.com/XuanMaoSecLab/DolphinChain/reproduce/nega_height_check	0.245s

```

## 修复

本漏洞相关修复见 : [Fix](https://github.com/tendermint/tendermint/commit/89cbcceac4d7359a4d0b38bedd137654279a006d)

修复方法：

```golang
    if blockHeight < 0 {
        return fmt.Errorf("Got a negative last block height (%d) from the app", blockHeight)
    }
```

## 相关资料

漏洞代码参考 : [未加验证](https://github.com/tendermint/tendermint/commit/89cbcceac4d7359a4d0b38bedd137654279a006d)

本漏洞相关 `Issue` 见 : [Issue](https://github.com/tendermint/tendermint/issues/911)
