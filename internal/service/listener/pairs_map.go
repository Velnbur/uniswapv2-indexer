package listener

import (
	"sync"

	uniswapv2pair "github.com/Velnbur/uniswapv2-indexer/contracts/uniswapv2-pair"
	"github.com/ethereum/go-ethereum/common"
)

type Pair struct {
	Address  common.Address
	Token0   common.Address
	Token1   common.Address
	Contract *uniswapv2pair.UniswapV2Pair
}

// PairsMap is a map of pairs.
type PairsMap struct {
	m sync.Map
}

// NewPairsMap returns a new PairsMap.
func NewPairsMap() *PairsMap {
	return &PairsMap{}
}

// Get returns the value for the given key.
func (m *PairsMap) Get(addr common.Address) *Pair {
	if v, ok := m.m.Load(addr); ok {
		return v.(*Pair)
	}
	return nil
}

// Set sets the value for the given key.
func (m *PairsMap) Set(addr common.Address, pair *Pair) {
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
func (m *PairsMap) Range(f func(key common.Address, value *Pair) bool) {
	m.m.Range(func(key, value interface{}) bool {
		return f(key.(common.Address), value.(*Pair))
	})
}
