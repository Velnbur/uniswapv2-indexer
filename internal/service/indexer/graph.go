package indexer

type Node struct {
	Symbol string

	weights map[string]map[string]int64
	routes  map[string]*Node
}

func NewNode(symbol string) *Node {
	return &Node{
		Symbol:  symbol,
		routes:  make(map[string]*Node),
		weights: make(map[string]map[string]int64),
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

func (g *Graph) AddPair(pair Pair) {
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

		for _, node := range g.nodes {
			if node.Symbol == symbol {
				continue
			}
			unvisited[node.Symbol] = node
		}

		weights := make(map[string]map[string]int64)
		for _, visitedNode := range visited {
			for _, edge := range visitedNode.routes {
				weights[visitedNode.Symbol][edge.Symbol] = visitedNode.weights[visitedNode.Symbol][edge.Symbol]
			}
		}
	}

}
