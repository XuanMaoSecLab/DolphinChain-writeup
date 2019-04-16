package block_for_loop

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	cfg "github.com/XuanMaoSecLab/DolphinChain/config"
	"github.com/XuanMaoSecLab/DolphinChain/libs/log"
	nm "github.com/XuanMaoSecLab/DolphinChain/node"

	"github.com/XuanMaoSecLab/DolphinChain/rpc/client"
)



func TestBlockchainInfoForloop(t *testing.T) {
	c := struct {
		min, max     int64
	}{
		-9223372036854775808, -9223372036854775788,
	}

	config := cfg.ResetTestRoot("node_node_test")
	defer os.RemoveAll(config.RootDir)
	n, err := nm.DefaultNewNode(config, log.TestingLogger())
	require.NoError(t, err)

	cl := client.NewLocal(n)
	t.Log(cl.NetInfo())
	cl.BlockchainInfo(c.min, c.max)
}
