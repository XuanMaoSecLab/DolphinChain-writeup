package make_fail

import (
	. "github.com/XuanMaoSecLab/DolphinChain/libs/common"
	"testing"
)


func TestNewBitArray(t *testing.T) {
	NewBitArray(-127)
}
