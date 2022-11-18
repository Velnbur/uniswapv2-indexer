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

func (wm *WeightsMap) Get(token0, token1 common.Address) (*big.Int, bool) {
	res, ok := wm.m.Load(weightsMapKey{token0, token1})
	if !ok {
		return nil, false
	}

	return res.(*big.Int), ok
}

func (wm *WeightsMap) Set(token0, token1 common.Address, weight *big.Int) {
	wm.m.Store(weightsMapKey{token0, token1}, weight)
}
