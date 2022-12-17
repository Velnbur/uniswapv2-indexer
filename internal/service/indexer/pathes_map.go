package indexer

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Velnbur/uniswapv2-indexer/internal/data"
)

type PathesMap struct {
	mutex sync.RWMutex

	m map[EdgeKey][]data.Path
}

func NewPathesMap() *PathesMap {
	return &PathesMap{
		m: make(map[EdgeKey][]data.Path),
	}
}

func (pm *PathesMap) AddPaths(pathes ...data.Path) {
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

		pm.m[key] = []data.Path{path}
	}
}

func (pm *PathesMap) GetPath(token0, token1 common.Address) []data.Path {
	if pathes, ok := pm.m[EdgeKey{token0, token1}]; ok {
		return pathes
	}

	return []data.Path{}
}
