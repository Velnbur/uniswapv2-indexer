package config

import (
	"github.com/Velnbur/uniswapv2-indexer/internal/contracts"
	"github.com/pkg/errors"

	contractsinit "github.com/Velnbur/uniswapv2-indexer/internal/config/contracts-init"
)

// FIXME: I don't know how sane this is to do, but it works for now to me
func (c *config) UniswapV2() *contracts.UniswapV2 {
	return c.uniswapv2.Do(func() interface{} {
		uniswapv2, err := contractsinit.InitUniswapV2(
			c.Log(), c.EthereumClient(), c.Redis(),
			c.ContracterCfg().Factory,
		)
		if err != nil {
			c.Log().Panic(errors.Wrap(err, "failed to init uniswapv2"))
		}

		return uniswapv2
	}).(*contracts.UniswapV2)
}
