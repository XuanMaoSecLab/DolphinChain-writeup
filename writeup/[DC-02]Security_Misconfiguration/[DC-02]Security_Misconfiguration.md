# 1. [DC-02] Security Misconfiguration

## 1.1. Tag

`config`;

`duplicateIP`

## 1.2. Vulnerability description

It's optional to allow duplicate IP in the default configure of the P2P network. The purpose of allowing duplicate IP is to make convenience for developers in their local debugging environment. However, allowing duplicate IP in the public network will bring risks.

## 1.3. Vulnerability analysis

File location : `config/config.go`

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

## 1.4. Reproduce

Here we use the testing script `security_misconfiguration_test.go` to test. You can create a file anywhere named `security_misconfiguration_test.go` and copy the content below, run it in the terminal.

### 1.4.1. Use go test script to test

```golang
// security_misconfiguration_test.go
package config

import (
    "testing"

    "github.com/stretchr/testify/assert"
    . "github.com/tendermint/tendermint/config"
)

// XuanMao: bug test
func TestP2PDefaultConfig(t *testing.T) {
    assert := assert.New(t)
    cfg := DefaultConfig()
    assert.NotNil(cfg.P2P)
    assert.False(cfg.P2P.AllowDuplicateIP)
}
```

### 1.4.2. Testing result

```sh
go test -v -run=TestP2PDefaultConfig
=== RUN   TestP2PDefaultConfig
--- FAIL: TestP2PDefaultConfig (0.00s)
    config_test.go:15: 
            Error Trace:    config_test.go:15
            Error:          Should be false
            Test:           TestP2PDefaultConfig
FAIL
FAIL    github.com/tendermint/tendermint/reproduce/sec_config    0.045s
```

## 1.5. Fix

Reference of this vulnerability: [Fix]((https://github.com/tendermint/tendermint/commit/24c1094ebcf2bd35f2642a44d7a1e5fb5c178fb1)).

### 1.5.1. Fix method

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
    AllowDuplicateIP:        false, // XuanMao: bug/risk fixed
    HandshakeTimeout:        20 * time.Second,
    DialTimeout:             3 * time.Second,
    TestDialFail:            false,
    TestFuzz:                false,
    TestFuzzConfig:          DefaultFuzzConnConfig(),
}
}
```

## 1.6. Reference

Vulnerable code from [set allow_duplicate_IP to false](https://github.com/tendermint/tendermint/commit/6108f7441f189c396ec4f346b8661175418fa2c9).

You can check related issue [here](https://github.com/tendermint/tendermint/issues/2712).