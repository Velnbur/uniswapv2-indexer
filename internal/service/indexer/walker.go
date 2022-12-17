package indexer

import (
	"github.com/Velnbur/uniswapv2-indexer/internal/data"
	"github.com/ethereum/go-ethereum/common"
)

type Walker struct {
	path data.Path

	root, current common.Address
}

func NewWalker(root common.Address) *Walker {
	return &Walker{
		root:    root,
		current: root,
		path:    data.NewPath(root),
	}
}

func (w *Walker) GetPath() []common.Address {
	return w.path
}

func (w *Walker) Current() common.Address {
	return w.current
}

func (w *Walker) Next(routes map[common.Address]*Edge) (bool, []*Walker) {
	walkers := make([]*Walker, 0)

	for next := range routes {
		if w.path.Contains(next) {
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
