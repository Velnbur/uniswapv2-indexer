package main

import (
	"context"
	"encoding/csv"
	"os"
	"strings"
	"sync"

	"github.com/alecthomas/kingpin"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"

	"github.com/Velnbur/uniswapv2-indexer/internal/contracts"
	"github.com/Velnbur/uniswapv2-indexer/internal/providers/inmemory"
	workerspool "github.com/Velnbur/uniswapv2-indexer/pkg/workers-pool"
)

var (
	// The output file
	output = kingpin.Flag("output", "Output file").
		Short('o').
		Default("output.csv").
		String()
	// Address of uniswapv2 factory
	factory = kingpin.Flag("factory", "Address of uniswapv2 factory").
		Short('f').
		Default("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f").
		String()
	// Ethereum node URI
	node = kingpin.Flag("node", "Node URI").
		Short('n').
		Default("https://cloudflare-eth.com/").
		String()

	loggerLevel = kingpin.Flag("log-level", "Log level").
			Short('l').
			Default("info").
			String()
)

// FIXME: this is a temporary solution, need to find a better way to handle rate
// limit errors that alchemy makes, when we make too many requests. May be,
// infura has better solutions for that
const errRateLimitStr = "Your app has exceeded its compute units"

func main() {
	kingpin.Parse()

	log := logan.New().
		WithField("service", "csv-stat")

	level, err := logan.ParseLevel(*loggerLevel)
	if err != nil {
		log.WithError(err).Fatal("failed to parse log level")
	}

	log = log.Level(level)

	client, err := ethclient.Dial(*node)
	if err != nil {
		log.WithError(err).Fatal("failed to connecto to ethereum node")
	}

	factoryProvider := inmemory.NewUniswapV2FactoryProvider()
	pairsProvider := inmemory.NewUniswapV2PairProvider()
	erc20Provider := inmemory.NewErc20Provider()

	factoryContract, err := contracts.NewUniswapV2Factory(
		common.HexToAddress(*factory),
		client, log,
		factoryProvider,
		pairsProvider,
		erc20Provider,
	)
	if err != nil {
		log.WithError(err).Fatal("failed to create factory contract")
	}

	file, err := os.OpenFile(*output, os.O_WRONLY|os.O_APPEND, os.ModeAppend)
	if err != nil {
		log.WithError(err).Fatal("failed to open output file")
	}

	mx := new(sync.Mutex)
	writer := csv.NewWriter(file)

	ctx := context.Background()

	amount, err := factoryContract.AllPairLength(ctx)
	if err != nil {
		log.WithError(err).Fatal("failed to get amount of pairs")
	}

	wp := workerspool.NewWorkingPool(50, int64(amount))

	for i := 0; i < int(amount); i++ {
		i := i
		wp.AddTask(func(ctx context.Context) error {
			f := func(ctx context.Context) error {
				pair, err := factoryContract.AllPairs(ctx, uint64(i))
				if err != nil {
					return errors.Wrap(err, "failed to get pair", logan.F{
						"index": i,
					})
				}
				token0, err := pair.Token0(ctx)
				if err != nil {
					return errors.Wrap(err, "failed to get token0", logan.F{
						"index": i,
					})
				}
				token1, err := pair.Token1(ctx)
				if err != nil {
					return errors.Wrap(err, "failed to get token1", logan.F{
						"index": i,
					})
				}
				reserve0, reserve1, err := pair.GetReserves(ctx)
				if err != nil {
					return errors.Wrap(err, "failed to get reserves", logan.F{
						"index": i,
					})
				}
				symbol0, err := token0.Symbol(ctx)
				if err != nil {
					return errors.Wrap(err, "failed to get symbol0", logan.F{
						"index": i,
					})
				}

				symbol1, err := token1.Symbol(ctx)
				if err != nil {
					return errors.Wrap(err, "failed to get symbol1", logan.F{
						"index": i,
					})
				}

				mx.Lock()
				defer mx.Unlock()
				err = writer.Write([]string{
					pair.Address.String(),
					token0.Address().Hex(),
					symbol0,
					reserve0.String(),
					token1.Address().Hex(),
					symbol1,
					reserve1.String(),
				})
				if err != nil {
					return errors.Wrap(err, "failed to write to file", logan.F{
						"index": i,
					})
				}

				writer.Flush()

				log.WithFields(logan.F{
					"index":   i,
					"pair":    pair.Address.String(),
					"symbol0": symbol0,
					"symbol1": symbol1,
				}).Debug("processed")
				return nil
			}

			if err := f(ctx); err != nil {
				if strings.Contains(err.Error(), errRateLimitStr) {
					return workerspool.RetryError
				}
				return err
			}
			return nil
		})
	}

	log.Info("starting to fetch pairs")

	if err := wp.Run(ctx); err != nil {
		log.WithError(err).Fatal("failed to run workers pool")
	}

	log.Info("done")
}
