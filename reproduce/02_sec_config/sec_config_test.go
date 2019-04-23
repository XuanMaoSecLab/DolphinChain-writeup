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
