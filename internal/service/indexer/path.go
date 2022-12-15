package indexer

import (
	"sync"

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

type PathesMap struct {
	mutex sync.RWMutex

	m map[EdgeKey][]Path
}

func NewPathesMap() *PathesMap {
	return &PathesMap{
		m: make(map[EdgeKey][]Path),
	}
}

func (pm *PathesMap) AddPaths(pathes ...Path) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	for _, path := range pathes {
		key := EdgeKey{path[0], path[len(path)-1]}
		keyInvers := EdgeKey{path[len(path)-1], path[0]}

		if _, ok := pm.m[key]; ok {
			pm.m[key] = append(pm.m[key], path)
			return
		}

		if _, ok := pm.m[keyInvers]; ok {
			pm.m[keyInvers] = append(pm.m[keyInvers], path)
			return
		}

		pm.m[key] = []Path{path}
	}
}

func (pm *PathesMap) GetPath(token0, token1 common.Address) []Path {
	if pathes, ok := pm.m[EdgeKey{token0, token1}]; ok {
		return pathes
	}

	return []Path{}
}
