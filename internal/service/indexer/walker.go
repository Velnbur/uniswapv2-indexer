package indexer

import "github.com/ethereum/go-ethereum/common"

type Walker struct {
	path Path

	root, current common.Address
	graph         *Graph
}

type Path map[common.Address]int

func NewPath(start common.Address) Path {
	return Path{start: 0}
}

func (p Path) GetSlice() []common.Address {
	path := make([]common.Address, len(p))
	for addr, i := range p {
		path[i] = addr
	}
	return path
}

func (p Path) Append(addr common.Address) {
	p[addr] = len(p)
}

func (p Path) Contains(addr common.Address) bool {
	_, ok := p[addr]
	return ok
}

func (p Path) Copy() Path {
	path := make(Path, len(p))
	for addr, i := range p {
		path[addr] = i
	}
	return path
}

func NewWalker(graph *Graph, root common.Address) *Walker {
	return &Walker{
		graph:   graph,
		root:    root,
		current: root,
		path:    map[common.Address]int{root: 0},
	}
}

func (w *Walker) GetPath() []common.Address {
	return w.path.GetSlice()
}

func (w *Walker) Next() (bool, []*Walker) {
	walkers := make([]*Walker, 0)

	for next := range w.graph.GetNode(w.current).routes {
		if w.path.Contains(next) {
			path := w.path.Copy()
			path.Append(next)

			walkers = append(walkers, &Walker{
				path:    path,
				root:    w.root,
				current: next,
				graph:   w.graph,
			})
		}
	}

	// If there are no more sub walkers (walker has no next routes),
	// we are done
	return len(walkers) > 0, walkers
}
