package indexer

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type weightsMapKey struct {
	Token0 common.Address
	Token1 common.Address
}

type weightsMapValue struct {
	Reserve0 *big.Int
	Reserve1 *big.Int
}

type WeightsMap struct {
	m sync.Map
}

func NewWeightsMap() *WeightsMap {
	return &WeightsMap{}
}

func (wm *WeightsMap) Get(token0, token1 common.Address) (*big.Int, *big.Int, bool) {
	res, ok := wm.m.Load(weightsMapKey{token0, token1})
	if !ok {
		return nil, nil, false
	}
	value := res.(weightsMapValue)

	reserve0 := new(big.Int).Set(value.Reserve0)
	reserve1 := new(big.Int).Set(value.Reserve1)

	return reserve0, reserve1, true
}

func (wm *WeightsMap) Set(
	token0, token1 common.Address,
	reserve0, reserve1 *big.Int,
) {
	value := weightsMapValue{
		Reserve0: new(big.Int).Set(reserve0),
		Reserve1: new(big.Int).Set(reserve1),
	}
	wm.m.Store(weightsMapKey{token0, token1}, value)
}
