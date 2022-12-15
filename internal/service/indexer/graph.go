package indexer

import (
	"math/big"
	"sync"

	"github.com/Velnbur/uniswapv2-indexer/pkg/math"
	"github.com/ethereum/go-ethereum/common"
)

type Graph struct {
	mux sync.RWMutex

	nodes map[common.Address]*Node
	edges map[common.Address]map[common.Address]*Edge

	pathesMap *PathesMap
}

func NewGraph() *Graph {
	return &Graph{
		// weights: make(map[EdgeKey]*big.Float),
		edges:     make(map[common.Address]map[common.Address]*Edge),
		nodes:     make(map[common.Address]*Node),
		pathesMap: NewPathesMap(),
	}
}

func (g *Graph) AddEdge(
	token0, token1 common.Address, reserve0, reserve1 *big.Int,
) *Graph {
	g.mux.Lock()
	defer g.mux.Unlock()

	edge := NewEdge(token0, token1, reserve0, reserve1)

	g.addEdge(edge)

	node0 := NewNode(token0)
	node1 := NewNode(token1)

	g.addNodes(node0, node1)

	return g
}

func (g *Graph) addEdge(edge *Edge) {
	if _, ok := g.edges[edge.Token0]; !ok {
		g.edges[edge.Token0] = make(map[common.Address]*Edge)
	}

	if _, ok := g.edges[edge.Token1]; !ok {
		g.edges[edge.Token1] = make(map[common.Address]*Edge)
	}

	g.edges[edge.Token0][edge.Token1] = edge
	g.edges[edge.Token1][edge.Token0] = edge
}

func (g *Graph) addNodes(nodes ...*Node) *Graph {
	for _, node := range nodes {
		g.nodes[node.Token] = node
	}

	return g
}

func (g *Graph) Index() {
	for node := range g.nodes {
		walkers := []*Walker{
			NewWalker(node),
		}
		pathes := make([]Path, 0)

		for len(walkers) > 0 {
			newWalkers := make([]*Walker, 0)

			for _, walker := range walkers {
				finished, _walkers := walker.Next(g.edges[walker.Current()])

				newWalkers = append(newWalkers, _walkers...)

				if finished {
					path := walker.GetPath()

					if len(path) > 1 {
						pathes = append(pathes, path)
					}
				}
			}

			walkers = newWalkers
		}

		g.pathesMap.AddPaths(pathes...)
	}
}

func (g *Graph) BestPath(input, ouput common.Address, amountIn *big.Int) (Path, *big.Int) {
	var (
		bestPath      = make(Path, 0)
		bestAmountOut = big.NewInt(0)
	)

	pathes := g.pathesMap.GetPath(input, ouput)

	for _, path := range pathes {
		reserves0 := make([]*big.Int, 0, len(path)-1)
		reserves1 := make([]*big.Int, 0, len(path)-1)

		for i := 1; i < len(path); i++ {
			edge := g.edges[input][ouput]

			reserves0 = append(reserves0, edge.Reserve0)
			reserves1 = append(reserves1, edge.Reserve1)
		}

		amountOut := math.SwapAmountOut(reserves0, reserves1, amountIn)

		if bestAmountOut.Cmp(amountOut) > 0 {
			bestAmountOut = amountOut
		}
	}

	return bestPath, bestAmountOut
}

func (g *Graph) UpdateReserves(
	token0, token1 common.Address, reserve0Delta, reserve1Delta *big.Int,
) {
	g.mux.Lock()
	defer g.mux.Unlock()

	edge := g.edges[token0][token1]
	edge.Reserve0.Add(edge.Reserve0, reserve0Delta)
	edge.Reserve1.Add(edge.Reserve1, reserve1Delta)
}
