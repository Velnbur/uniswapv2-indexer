package contracts

import (
	"context"

	"github.com/Velnbur/uniswapv2-indexer/generated/erc20"
	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

type ERC20 struct {
	contract *erc20.Erc20

	symbol   string
	address  common.Address
	provider providers.Erc20Provider
}

func NewERC20(
	address common.Address, client *ethclient.Client, provider providers.Erc20Provider,
) (*ERC20, error) {
	contract, err := erc20.NewErc20(address, client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create erc20 contract")
	}

	return &ERC20{
		contract: contract,
		address:  address,
		provider: provider,
	}, nil
}

func (e *ERC20) Address() common.Address {
	return e.address
}

func (e *ERC20) Symbol(ctx context.Context) ([]byte, error) {
	if e.symbol != "" {
		return e.symbol, nil
	}

	symbol, err := e.provider.GetSymbol(ctx, e.address)
	if err != nil {
		return "", errors.Wrap(err, "failed to get symbol from cache")
	}
	if symbol != "" {
		e.symbol = symbol
		return symbol, nil
	}

	symbol, err = e.contract.Symbol(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to get symbol from contract")
	}

	err = e.provider.SetSymbol(ctx, e.address, symbol)
	if err != nil {
		return "", errors.Wrap(err, "failed to set symbol")
	}

	e.symbol = symbol
	return symbol, nil
}
