package indexer

import (
	"github.com/ethereum/go-ethereum/common"
)

type Node struct {
	Token  common.Address
	Symbol string
}

func NewNode(addr common.Address) *Node {
	return &Node{
		Token: addr,
	}
}

func (n *Node) AddSymbol(symbol string) *Node {
	n.Symbol = symbol
	return n
}
