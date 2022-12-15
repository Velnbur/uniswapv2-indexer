package config

import (
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"

	"github.com/Velnbur/uniswapv2-indexer/internal/contracts"
	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
	"github.com/ethereum/go-ethereum/common"
)

const yamlTokensKey = "tokens"

func (c *config) Tokens() []*contracts.ERC20 {
	return c.tokens.Do(func() interface{} {
		tokensMap := make(map[string]string)

		err := figure.Out(&tokensMap).
			From(kv.MustGetStringMap(c.getter, yamlTokensKey)).
			Please()
		if err != nil {
			c.Log().WithError(err).Panic("failed to parse config")
		}

		erc20Tokens := make([]*contracts.ERC20, 0)

		for name, token := range tokensMap {
			erc20, err := contracts.NewERC20(contracts.Erc20Config{
				Address:  common.HexToAddress(token),
				Client:   c.EthereumClient(),
				Provider: providers.NewErc20RedisProvider(c.Redis()),
			})
			if err != nil {
				c.Log().
					WithError(err).
					WithFields(logan.F{
						"address": token,
						"token":   name,
					}).Panic("failed to init erc20 cotract")
			}

			erc20Tokens = append(erc20Tokens, erc20)
		}

		return erc20Tokens
	}).([]*contracts.ERC20)
}
