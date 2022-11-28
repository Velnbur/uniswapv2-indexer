package inmemory

import (
	"context"
	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var _ providers.UniswapV2PairProvider = &UniswapV2PairProvider{}

type uniswapV2PairValue struct {
	Token0, Token1       common.Address
	Reserves0, Reserves1 *big.Int
}

type UniswapV2PairProvider struct {
	pairs map[common.Address]uniswapV2PairValue
}

func NewUniswapV2PairProvider() *UniswapV2PairProvider {
	return &UniswapV2PairProvider{
		pairs: make(map[common.Address]uniswapV2PairValue),
	}
}

func (p *UniswapV2PairProvider) GetReserves(
	_ context.Context, pair common.Address,
) (reserves0, reserves1 *big.Int, err error) {
	pairValue, ok := p.pairs[pair]
	if !ok {
		return nil, nil, nil
	}

	return pairValue.Reserves0, pairValue.Reserves1, nil
}

func (p *UniswapV2PairProvider) GetTokens(
	_ context.Context, pair common.Address,
) (token0, token1 common.Address, err error) {
	pairValue, ok := p.pairs[pair]
	if !ok {
		return common.Address{}, common.Address{}, nil
	}

	return pairValue.Token0, pairValue.Token1, nil
}

// SetReserves implements providers.UniswapV2PairProvider
func (p *UniswapV2PairProvider) SetReserves(
	_ context.Context, pair common.Address, reserve0, reserve1 *big.Int,
) error {
	pairValue := p.pairs[pair]

	pairValue.Reserves0 = reserve0
	pairValue.Reserves1 = reserve1

	p.pairs[pair] = pairValue
	return nil
}

// SetTokens implements providers.UniswapV2PairProvider
func (p *UniswapV2PairProvider) SetTokens(
	_ context.Context, pair common.Address, token0, token1 common.Address,
) error {
	pairValue := p.pairs[pair]

	pairValue.Token0 = token0
	pairValue.Token1 = token1

	p.pairs[pair] = pairValue
	return nil
}
