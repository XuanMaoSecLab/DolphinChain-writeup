package hrd_validate

import (
	"testing"
	"time"
	"bytes"

//	"github.com/XuanMaoSecLab/DolphinChain/lite"
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
