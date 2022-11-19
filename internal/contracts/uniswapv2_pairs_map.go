package contracts

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

// UniswapV2PairsMap is a map of pairs.
type UniswapV2PairsMap struct {
	m sync.Map
}

// NewPairsMap returns a new PairsMap.
func NewPairsMap() *UniswapV2PairsMap {
	return &UniswapV2PairsMap{}
}

// Get returns the value for the given key.
func (m *UniswapV2PairsMap) Get(addr common.Address) *UniswapV2Pair {
	if v, ok := m.m.Load(addr); ok {
		return v.(*UniswapV2Pair)
	}
	return nil
}

// Set sets the value for the given key.
func (m *UniswapV2PairsMap) Set(addr common.Address, pair *UniswapV2Pair) {
	m.m.Store(addr, pair)
}

func (m *UniswapV2PairsMap) Delete(addr common.Address) {
	m.m.Delete(addr)
}

func (m *UniswapV2PairsMap) Len() int {
	var count int
	m.m.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// Range calls f sequentially for each key and value present in the map.
func (m *UniswapV2PairsMap) Range(f func(key common.Address, value *UniswapV2Pair) bool) {
	m.m.Range(func(key, value interface{}) bool {
		return f(key.(common.Address), value.(*UniswapV2Pair))
	})
}
