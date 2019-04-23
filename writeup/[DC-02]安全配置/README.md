# [DC-02] 安全配置

## 漏洞标签

`config`;

`duplicateIP`

## 漏洞描述

安全配置(Security Misconfiguration)

P2P网络默认设置里可选允许重复IP或禁止重复IP。允许重复IP是为了方便开发者在本地调试。但是如果在公网上允许重复IP会带来风险

## 漏洞分析

文件：`config/config.go`

```golang
// config/config.go

func DefaultP2PConfig() *P2PConfig {
return &P2PConfig{
    ListenAddress:           "tcp://0.0.0.0:26656",
    ExternalAddress:         "",
    UPNP:                    false,
    AddrBook:                defaultAddrBookPath,
    AddrBookStrict:          true,
    MaxNumInboundPeers:      40,
    MaxNumOutboundPeers:     10,
    FlushThrottleTimeout:    100 * time.Millisecond,
    MaxPacketMsgPayloadSize: 1024,    // 1 kB
    SendRate:                5120000, // 5 mB/s
    RecvRate:                5120000, // 5 mB/s
    PexReactor:              true,
    SeedMode:                false,
    AllowDuplicateIP:        true, // XuanMao : risk here
    HandshakeTimeout:        20 * time.Second,
    DialTimeout:             3 * time.Second,
    TestDialFail:            false,
    TestFuzz:                false,
    TestFuzzConfig:          DefaultFuzzConnConfig(),
}
}

```

## 复现或测试步骤

此处使用test脚本测试

### 使用 go test 脚本测试

测试脚本位于 `DolphinChain-writeup\reproduce` 中。

```golang
package config

import (
    "testing"

    "github.com/stretchr/testify/assert"
    . "github.com/XuanMaoSecLab/DolphinChain/config"
)

// XuanMao : bug test
func TestP2PDefaultConfig(t *testing.T) {
    assert := assert.New(t)
    cfg := DefaultConfig()
    assert.NotNil(cfg.P2P)
    assert.False(cfg.P2P.AllowDuplicateIP)
}
```

运行结果

```sh
go test -v -run=TestP2PDefaultConfig
=== RUN   TestP2PDefaultConfig
--- FAIL: TestP2PDefaultConfig (0.00s)
    config_test.go:15: 
        	Error Trace:	config_test.go:15
        	Error:      	Should be false
        	Test:       	TestP2PDefaultConfig
FAIL
FAIL	github.com/XuanMaoSecLab/DolphinChain/reproduce/sec_config	0.045s
```

## 修复

本漏洞相关修复见 : [Fix](https://github.com/tendermint/tendermint/commit/24c1094ebcf2bd35f2642a44d7a1e5fb5c178fb1)

修复方法：

将默认配置的`AllowDuplicateIP`改为`false`，同时增加一个测试配置函数`TestP2PConfig()`，`AllowDuplicateIP`默认为`true`。

```golang
func DefaultP2PConfig() *P2PConfig {
return &P2PConfig{
    ListenAddress:           "tcp://0.0.0.0:26656",
    ExternalAddress:         "",
    UPNP:                    false,
    AddrBook:                defaultAddrBookPath,
    AddrBookStrict:          true,
    MaxNumInboundPeers:      40,
    MaxNumOutboundPeers:     10,
    FlushThrottleTimeout:    100 * time.Millisecond,
    MaxPacketMsgPayloadSize: 1024,    // 1 kB
    SendRate:                5120000, // 5 mB/s
    RecvRate:                5120000, // 5 mB/s
    PexReactor:              true,
    SeedMode:                false,
    AllowDuplicateIP:        false, // XuanMao: bug fixed
    HandshakeTimeout:        20 * time.Second,
    DialTimeout:             3 * time.Second,
    TestDialFail:            false,
    TestFuzz:                false,
    TestFuzzConfig:          DefaultFuzzConnConfig(),
}
}
```

```golang
// XunaMao: fixed, add Test Config
func TestP2PConfig() *P2PConfig {
    cfg := DefaultP2PConfig()
    cfg.ListenAddress = "tcp://0.0.0.0:36656"
    cfg.FlushThrottleTimeout = 10 * time.Millisecond
    cfg.AllowDuplicateIP = true
    return cfg
}


// TestConfig returns a configuration that can be used for testing
func TestConfig() *Config {
    return &Config{
        BaseConfig:      TestBaseConfig(),
        RPC:             TestRPCConfig(),
        P2P:             TestP2PConfig(), // XuanMao: bug fixed
        Mempool:         TestMempoolConfig(),
        Consensus:       TestConsensusConfig(),
        TxIndex:         TestTxIndexConfig(),
        Instrumentation: TestInstrumentationConfig(),
    }
}
```

## 相关资料

漏洞代码参考: [设置allow_duplicate_ip为False](https://github.com/tendermint/tendermint/commit/6108f7441f189c396ec4f346b8661175418fa2c9)

本漏洞相关 `Issue` 见 : [Issue](https://github.com/tendermint/tendermint/issues/2712)
