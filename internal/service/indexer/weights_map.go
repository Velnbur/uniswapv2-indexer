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

type WeightsMap struct {
	m sync.Map
}

func NewWeightsMap() *WeightsMap {
	return &WeightsMap{}
}

func (wm *WeightsMap) Get(token0, token1 common.Address) (*big.Float, bool) {
	res, ok := wm.m.Load(weightsMapKey{token0, token1})
	if !ok {
		return nil, false
	}
	value := res.(*big.Float)

	return new(big.Float).Set(value), true
}

func (wm *WeightsMap) Set(token0, token1 common.Address, value *big.Float) {
	wm.m.Store(weightsMapKey{token0, token1}, new(big.Float).Set(value))
}
