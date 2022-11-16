package main

import (
	"os"

	"github.com/Velnbur/uniswapv2-indexer/internal/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}
