# 1. [DC-08] Inconsistence of Message Size Limit in Mempool

## 1.1. Tag

`Tx_check_size`;

`Msg_check_size`;

`Mem_pool`

## 1.2. Vulnerability description

Inconsistence of Message Size Limit in Mempool

memPool should check the transaction when it accepts the transaction(checkTx) from the nodes.
Meanwhile, when other nodes(or validators) accept the transactions from the mempool, they will also check the transactions(checkMsg). It's important to notice that the size of the transaction should keep consistency.

The vulnerability lies here is the mempool doesn't check the size of transactions, which lead the failure of verification for other nodes mistakenly accept wrong size transactions and disconnection with mempool.

## 1.3. Vulnerability analysis

Vulnerable code: [mempool validate the transaction](https://github.com/tendermint/tendermint/blob/8003786c9affff242861141bf7484aeb5796e42c/mempool/mempool.go#L303-L348)

[node decode the message](https://github.com/tendermint/tendermint/blob/8003786c9affff242861141bf7484aeb5796e42c/mempool/reactor.go#L184-L190)

when mempool accept the transaction, it will check the transaction.

```golang
// mempool check the transaction
func (mem *Mempool) CheckTx(tx types.Tx, cb func(*abci.Response)) (err error) { 
     mem.proxyMtx.Lock() 
     // use defer to unlock mutex because application (*local client*) might panic 
     defer mem.proxyMtx.Unlock() 
  
     if mem.Size() >= mem.config.Size { 
         return ErrMempoolIsFull 
     } 
  
     if mem.preCheck != nil { 
         if err := mem.preCheck(tx); err != nil { 
             return ErrPreCheck{err} 
         } 
     } 
  
     // CACHE 
     if !mem.cache.Push(tx) { 
         return ErrTxInCache 
     } 
     // END CACHE 
  
     // WAL 
     if mem.wal != nil { 
         // TODO: Notify administrators when WAL fails 
         _, err := mem.wal.Write([]byte(tx)) 
         if err != nil { 
             mem.logger.Error("Error writing to WAL", "err", err) 
         } 
         _, err = mem.wal.Write([]byte("\n")) 
         if err != nil { 
             mem.logger.Error("Error writing to WAL", "err", err) 
         } 
     } 
     // END WAL 
  
     // NOTE: proxyAppConn may error if tx buffer is full 
     if err = mem.proxyAppConn.Error(); err != nil { 
         return err 
     } 
     reqRes := mem.proxyAppConn.CheckTxAsync(tx) 
     if cb != nil { 
         reqRes.SetCallback(cb) 
     } 
  
     return nil 
 } 
```

when other nodes(or validators) accept the transactions from the mempool, they will also check the transactions(checkMsg)

```go
 func decodeMsg(bz []byte) (msg MempoolMessage, err error) { 
     if len(bz) > maxMsgSize { 
         return msg, fmt.Errorf("Msg exceeds max size (%d > %d)", len(bz), maxMsgSize) 
     } 
     err = cdc.UnmarshalBinaryBare(bz, &msg) 
     return 
 } 
```

## 1.4. Reproduce

Here we use test script to test, copy the content of script below to `XuanMaoSecLab/DolphinChain/mempool`, name the script as `mem_msg_test.go`. Here is only a demo, the actual impact of disconnection doesn't show here.

### 1.4.1. Use go test script to test

```golang
// XuanMao: bug test
// mem_msg_test.go
package mempool

import (
    "fmt"
    "testing"

    "github.com/stretchr/testify/require"


    "github.com/XuanMaoSecLab/DolphinChain/abci/example/kvstore"
    cmn "github.com/XuanMaoSecLab/DolphinChain/libs/common"
    "github.com/XuanMaoSecLab/DolphinChain/proxy"
)


func TestCheckDemo(t *testing.T) {
    app := kvstore.NewKVStoreApplication()
    cc := proxy.NewLocalClientCreator(app)
    mempl, cleanup := newMempoolWithApp(cc)
    defer cleanup()

    testCases := []struct {
        len int
        err bool
    }{
        // check small txs. no error
        {10, false},
        {1000, false},
        {1000000, false},

        // check around maxTxSize
        // changes from no error to error
        {maxTxSize - 2, false},
        {maxTxSize - 1, false},
        {maxTxSize, false},
        {maxTxSize + 1, true},
        {maxTxSize + 2, true},

        // check around maxMsgSize. all error
        {maxMsgSize - 1, true},
        {maxMsgSize, true},
        {maxMsgSize + 1, true},
    }

    for i, testCase := range testCases {
        caseString := fmt.Sprintf("case %d, len %d", i, testCase.len)
        tx := cmn.RandBytes(testCase.len)
        err := mempl.CheckTx(tx, nil)
        msg := &TxMessage{tx}
        //cdc := Getcdc()
        encoded := cdc.MustMarshalBinaryBare(msg)
        _ , err_d := decodeMsg(encoded)
        //t.Log(msg1)
        if err_d != nil {
            t.Error(err_d)    
        }
        require.Equal(t, len(encoded), txMessageSize(tx), caseString)
        if !testCase.err {
            require.True(t, len(encoded) <= maxMsgSize, caseString)
            require.NoError(t, err, caseString)
        } else {
            require.True(t, len(encoded) > maxMsgSize, caseString)
            //require.Equal(t, err, ErrTxTooLarge, caseString)
        }
    }

}

```

### 1.4.2. Testing result

We can see we add 11 transactions in total, in fact, the last 5 transactions should not be added in the normal situation.

```sh
[root@MiWiFi-R3P-srv mempool]# go test -v -run=TestCheckDemo
=== RUN   TestCheckDemo
I[2019-04-02|15:59:21.076] Starting localClient                         module=abci-client connection=mempool impl=localClient
I[2019-04-02|15:59:21.078] Added good transaction                       tx=C18C2FBA225BF71FEF81D34D066FDFFD2DD8F5AA2F4D9619CD73B0E9AC10C247 res="&{CheckTx:gas_wanted:1 }" height=0 total=1
I[2019-04-02|15:59:21.080] Added good transaction                       tx=8313B25C160687B31113E872453108939A6D392B31E52BEC5CB11015E56ACBD6 res="&{CheckTx:gas_wanted:1 }" height=0 total=2
I[2019-04-02|15:59:21.178] Added good transaction                       tx=A819541F0EC5912675B90EB2C71E3CB3E504FE0E1823C2FD37B7BA4CBAEB5E9B res="&{CheckTx:gas_wanted:1 }" height=0 total=3
I[2019-04-02|15:59:21.293] Added good transaction                       tx=2F6C8FCCA3903F74057D888CD548064CB58676EAE183F8D20F69257843290065 res="&{CheckTx:gas_wanted:1 }" height=0 total=4
I[2019-04-02|15:59:21.410] Added good transaction                       tx=96A4E34F1D8670214A4CEC9D383DC52266D0CDE3EA0A8304585D26DDC049649D res="&{CheckTx:gas_wanted:1 }" height=0 total=5
I[2019-04-02|15:59:21.501] Added good transaction                       tx=550B12D2083F612A1195181062D3B1B8E12DED905D2AEBC9EB4F6C0C4C6814CC res="&{CheckTx:gas_wanted:1 }" height=0 total=6
I[2019-04-02|15:59:21.608] Added good transaction                       tx=A874320438EAA0B499A70D79627FDBA577DD07FEA426AD896D49B74E161F9760 res="&{CheckTx:gas_wanted:1 }" height=0 total=7
I[2019-04-02|15:59:21.691] Added good transaction                       tx=44784BA98CBD9654FF593D0B04BB3BA2730050BBB157D7C7F41796558B33CB57 res="&{CheckTx:gas_wanted:1 }" height=0 total=8
I[2019-04-02|15:59:21.792] Added good transaction                       tx=9695184C6B09BB6A96EC436AF5638596271EE5C4556AB2F23F1F124816329765 res="&{CheckTx:gas_wanted:1 }" height=0 total=9
I[2019-04-02|15:59:21.901] Added good transaction                       tx=A0CA845CBF81DCE22569DC1BDC698126BBE49A21E008B6A4FB87CDF5A02537AA res="&{CheckTx:gas_wanted:1 }" height=0 total=10
I[2019-04-02|15:59:21.993] Added good transaction                       tx=053273DCA4783009EC3CA80C080EAF705EF0AF737537288DA8960114D7984CB4 res="&{CheckTx:gas_wanted:1 }" height=0 total=11
--- FAIL: TestCheckDemo (0.92s)
    mem_msg_test.go:55: Msg exceeds max size (1048577 > 1048576)
    mem_msg_test.go:55: Msg exceeds max size (1048578 > 1048576)
    mem_msg_test.go:55: Msg exceeds max size (1048583 > 1048576)
    mem_msg_test.go:55: Msg exceeds max size (1048584 > 1048576)
    mem_msg_test.go:55: Msg exceeds max size (1048585 > 1048576)
FAIL
exit status 1
FAIL    github.com/XuanMaoSecLab/DolphinChain/mempool    0.992s


```

## 1.5. Fix

Reference of this vulnerability: [Fix](https://github.com/tendermint/tendermint/commit/da95f4aa6da2b966fe9243e481e6cfb3bf3b2c5a)

The vulnerability has been fixed in `da95f4a`.

### 1.5.1. Fix method

```golang
// add verification in /mempool/mempool.go
func (mem *Mempool) CheckTx(tx types.Tx, cb func(*abci.Response)) (err error) {
    ...
        if len(tx) > maxTxSize {
         return ErrTxTooLarge
     }
     ...
```

After fixing, let's run the testing script, we can see only first 6 valid transactions had been added, the last 5 invalid transactions was not added into the mempool.

```sh
[root@MiWiFi-R3P-srv mempool]# go test -v run=TestCheckDemo
=== RUN   TestCheckDemo
I[2019-04-02|16:01:07.501] Starting localClient                         module=abci-client connection=mempool impl=localClient
I[2019-04-02|16:01:07.502] Added good transaction                       tx=92CBA29FC15AD522058F9A777438EEF0B3713023F0A14982E17F3C3663344BAE res="&{CheckTx:gas_wanted:1 }" height=0 total=1
I[2019-04-02|16:01:07.504] Added good transaction                       tx=D4A9B96F9A5E529FE68050CE9CAAEF5CD2076457D23BE0BF8F2821899BAC2379 res="&{CheckTx:gas_wanted:1 }" height=0 total=2
I[2019-04-02|16:01:07.603] Added good transaction                       tx=C0B45A9AB67CBB457C178ADFCE1FD2C81005D512381D6AD4E8BC6D2A1AABA318 res="&{CheckTx:gas_wanted:1 }" height=0 total=3
I[2019-04-02|16:01:07.710] Added good transaction                       tx=8A7106FA021BB958A88C500A7F143EFE76DDCAEBA4379A0F16FB0D881650DB26 res="&{CheckTx:gas_wanted:1 }" height=0 total=4
I[2019-04-02|16:01:07.810] Added good transaction                       tx=53B3F4363FFD1DB89DBD36DF2C1A90B989A7CD00AD6CED2AE55FFC616FC6883B res="&{CheckTx:gas_wanted:1 }" height=0 total=5
I[2019-04-02|16:01:07.919] Added good transaction                       tx=01FC2C215A246D87174A34F6ED907520835869CC677DAEFE32A444FCA733B380 res="&{CheckTx:gas_wanted:1 }" height=0 total=6
--- FAIL: TestCheckDemo (0.76s)
    mem_msg_test.go:55: Msg exceeds max size (1048577 > 1048576)
    mem_msg_test.go:55: Msg exceeds max size (1048578 > 1048576)
    mem_msg_test.go:55: Msg exceeds max size (1048583 > 1048576)
    mem_msg_test.go:55: Msg exceeds max size (1048584 > 1048576)
    mem_msg_test.go:55: Msg exceeds max size (1048585 > 1048576)
FAIL
exit status 1
FAIL    github.com/XuanMaoSecLab/DolphinChain/mempool    0.829s

```

## 1.6. Reference

You can check related issue [here](https://github.com/tendermint/tendermint/issues/3008)
