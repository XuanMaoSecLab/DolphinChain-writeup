package clist

import (
	"testing"

	. "github.com/XuanMaoSecLab/DolphinChain/libs/clist"
)

func TestPanicOnMaxLength(t *testing.T) {
	t.Log(MaxLength)
	l := new(CList)
	l.Init()
	for i := 0; i < MaxLength; i++ {
		l.PushBack(100)
	}
}
