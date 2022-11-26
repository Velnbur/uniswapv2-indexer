package indexer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Node struct {
	Token common.Address

	weights *WeightsMap
	routes  map[common.Address]*Node
}

func NewNode(addr common.Address) *Node {
	return &Node{
		Token:   addr,
		routes:  make(map[common.Address]*Node),
		weights: NewWeightsMap(),
	}
}

type Graph struct {
	nodes map[common.Address]*Node
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[common.Address]*Node),
	}
}

func (g *Graph) GetNode(token common.Address) *Node {
	return g.nodes[token]
}

func (g *Graph) AddPair(
	token0, token1 common.Address,
	reserve0, reserve1 *big.Int,
) {
	node0, ok := g.nodes[token0]
	if !ok {
		node0 = NewNode(token0)
		g.nodes[token0] = node0
	}
	node1, ok := g.nodes[token1]
	if !ok {
		node1 = NewNode(token1)
		g.nodes[token1] = node1
	}
	node1.weights.Set(token1, token0, reserve1, reserve0)
	node0.weights.Set(token0, token1, reserve0, reserve1)
	node0.routes[token1] = node1
	node1.routes[token0] = node0
}

func (g *Graph) indexAllPathes() {
	var pathes [][]common.Address

	for _, node := range g.nodes {
		pathes = append(pathes, g.indexPathes(node)...)
	}

	for _, path := range pathes {
		g.calculateWieghtsForPath(path)
	}
}

func (g *Graph) indexPathes(node *Node) [][]common.Address {
	var pathes [][]common.Address
	walkers := []*Walker{
		NewWalker(g, node.Token),
	}

	for _, walker := range walkers {
		done, nextWalkers := walker.Next()
		if done {
			pathes = append(pathes, walker.GetPath())
		}
		walkers = append(walkers, nextWalkers...)
	}

	return pathes
}

func (g *Graph) calculateWieghtsForPath(path []common.Address) {
	for i := 0; i < len(path)-1; i++ {
		// TODO:
	}
}
