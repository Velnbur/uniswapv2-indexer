package contracts

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"

	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
)

type UniswapV2 struct {
	Factory *UniswapV2Factory
	Pairs   *UniswapV2PairsMap
}

func NewUniswapV2(
	factoryAddr common.Address, client *ethclient.Client, logger *logan.Entry,
	factoryProvider providers.UniswapV2FactoryProvider,
	pairProvider providers.UniswapV2PairProvider,
	erc20 providers.Erc20Provider,
) (*UniswapV2, error) {
	factory, err := NewUniswapV2Factory(UniswapV2FactoryConfig{
		factoryAddr, client, logger,
		factoryProvider, pairProvider, erc20,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create UniswapV2Factory")
	}
	return &UniswapV2{
		Factory: factory,
		Pairs:   NewPairsMap(),
	}, nil
}
