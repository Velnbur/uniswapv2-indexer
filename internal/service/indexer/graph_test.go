package indexer

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func Test_GraphIndex(t *testing.T) {
	t.Run("index graph with 5 nodes and 10 edges", func(t *testing.T) {
		tokens := []common.Address{
			common.HexToAddress("0x0000000000000000000000000000000000000001"),
			common.HexToAddress("0x0000000000000000000000000000000000000002"),
			common.HexToAddress("0x0000000000000000000000000000000000000003"),
			common.HexToAddress("0x0000000000000000000000000000000000000004"),
			common.HexToAddress("0x0000000000000000000000000000000000000005"),
		}

		graph := NewGraph()

		graph.
			AddEdge(tokens[0], tokens[1], big.NewInt(0), big.NewInt(0)).
			AddEdge(tokens[0], tokens[2], big.NewInt(0), big.NewInt(0)).
			AddEdge(tokens[0], tokens[3], big.NewInt(0), big.NewInt(0)).
			AddEdge(tokens[0], tokens[4], big.NewInt(0), big.NewInt(0)).
			AddEdge(tokens[1], tokens[2], big.NewInt(0), big.NewInt(0)).
			AddEdge(tokens[1], tokens[3], big.NewInt(0), big.NewInt(0)).
			AddEdge(tokens[1], tokens[4], big.NewInt(0), big.NewInt(0)).
			AddEdge(tokens[2], tokens[3], big.NewInt(0), big.NewInt(0)).
			AddEdge(tokens[2], tokens[4], big.NewInt(0), big.NewInt(0)).
			AddEdge(tokens[3], tokens[4], big.NewInt(0), big.NewInt(0))

		graph.Index()
	})
}
