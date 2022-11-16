package indexer

import "testing"

func Test_Graph(t *testing.T) {
	t.Run("Add pairs", func(t *testing.T) {
		graph := NewGraph()

		graph.AddPair(Pair{Token1: "USDT", Token2: "BNB"})
		graph.AddPair(Pair{Token1: "BNB", Token2: "MMPRO"})
		graph.AddPair(Pair{Token1: "MMPRO", Token2: "BNB"})
	})
}
