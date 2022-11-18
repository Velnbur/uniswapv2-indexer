package listener

import (
	"sync"

	"github.com/Velnbur/uniswapv2-indexer/internal/contracts"
	"github.com/ethereum/go-ethereum/common"
)

// PairsMap is a map of pairs.
type PairsMap struct {
	m sync.Map
}

// NewPairsMap returns a new PairsMap.
func NewPairsMap() *PairsMap {
	return &PairsMap{}
}

// Get returns the value for the given key.
func (m *PairsMap) Get(addr common.Address) *contracts.UniswapV2Pair {
	if v, ok := m.m.Load(addr); ok {
		return v.(*contracts.UniswapV2Pair)
	}
	return nil
}

// Set sets the value for the given key.
func (m *PairsMap) Set(addr common.Address, pair *contracts.UniswapV2Pair) {
	m.m.Store(addr, pair)
}

func (m *PairsMap) Delete(addr common.Address) {
	m.m.Delete(addr)
}

func (m *PairsMap) Len() int {
	var count int
	m.m.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// Range calls f sequentially for each key and value present in the map.
func (m *PairsMap) Range(f func(key common.Address, value *contracts.UniswapV2Pair) bool) {
	m.m.Range(func(key, value interface{}) bool {
		return f(key.(common.Address), value.(*contracts.UniswapV2Pair))
	})
}
