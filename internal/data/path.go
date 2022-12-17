package data

import (
	"github.com/ethereum/go-ethereum/common"
)

type Path []common.Address

func NewPath(start common.Address) Path {
	return Path{start}
}

func (p Path) Append(addr common.Address) {
	p = append(p, addr)
}

func (p Path) Contains(addr common.Address) bool {
	for _, elem := range p {
		if elem == addr {
			return true
		}
	}

	return false
}

func (p Path) Copy() Path {
	path := make(Path, len(p))

	copy(path, p)

	return path
}
