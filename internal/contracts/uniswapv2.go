package contracts

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"gitlab.com/distributed_lab/logan/v3"
)

type UniswapV2 struct {
	Factory *UniswapV2Factory
	Pairs   *UniswapV2PairsMap
}

func NewUniswapV2(
	factoryAddr common.Address, client *ethclient.Client, redis *redis.Client,
	logger *logan.Entry,
) (*UniswapV2, error) {
	factory, err := NewUniswapV2Factory(factoryAddr, client, redis, logger)
	if err != nil {
		return nil, err
	}
	return &UniswapV2{
		Factory: factory,
		Pairs:   NewPairsMap(),
	}, nil
}
