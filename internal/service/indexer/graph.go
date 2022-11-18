package indexer

import (
	"github.com/Velnbur/uniswapv2-indexer/internal/contracts"
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
	nodes map[string]*Node
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
	}
}

func (g *Graph) AddPair(pair *contracts.UniswapV2Pair) {
	node, ok := g.nodes[pair.Token1]
	if !ok {
		node = NewNode(pair.Token1)
		g.nodes[pair.Token1] = node
	}

	node.routes[pair.Token2] = NewNode(pair.Token2)
}

func (g *Graph) indexWeights() {
	for symbol, node := range g.nodes {
		visited := make(map[string]*Node, len(g.nodes))
		visited[symbol] = node
		unvisited := make(map[string]*Node, len(g.nodes)-1)

	}

}
