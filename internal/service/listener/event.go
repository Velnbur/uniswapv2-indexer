package listener

import (
	uniswapv2factory "github.com/Velnbur/uniswapv2-indexer/contracts/uniswapv2-factory"
	uniswapv2pair "github.com/Velnbur/uniswapv2-indexer/contracts/uniswapv2-pair"
)

type EventType int

const (
	EventTypeSwap EventType = iota
	EventTypeMint
	EventTypeBurn
	EventTypePairCreated
)

type Event struct {
	Type       EventType
	Burn       *uniswapv2pair.UniswapV2PairBurn
	Swap       *uniswapv2pair.UniswapV2PairSwap
	Mint       *uniswapv2pair.UniswapV2PairMint
	CreatePair *uniswapv2factory.UniswapV2FactoryPairCreated
}
