package indexer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Walker struct {
	path Path

	root, current common.Address
}

func NewWalker(root common.Address) *Walker {
	return &Walker{
		root:    root,
		current: root,
		path:    NewPath(root),
	}
}

func (w *Walker) GetPath() []common.Address {
	return w.path
}

func (w *Walker) Current() common.Address {
	return w.current
}

const MinimumLiquidity = 1000

var (
	minimumLiquidityBig = big.NewInt(MinimumLiquidity)
)

func (w *Walker) Next(routes map[common.Address]*Edge) (bool, []*Walker) {
	walkers := make([]*Walker, 0)

	for next, edge := range routes {
		if w.path.Contains(next) {
			continue
		}

		// this route is not  usable bacause lack of liquidity
		if edge.Reserve0.Cmp(minimumLiquidityBig) < 0 || edge.Reserve1.Cmp(minimumLiquidityBig) < 0 {
			continue
		}

		path := w.path.Copy()
		path.Append(next)

		walkers = append(walkers, &Walker{
			path:    path,
			root:    w.root,
			current: next,
		})
	}

	// If there are no more sub walkers (walker has no next routes),
	// we are done
	return len(walkers) > 0, walkers
}
