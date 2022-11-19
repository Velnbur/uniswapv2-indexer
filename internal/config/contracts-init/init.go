package contractsinit

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"

	"github.com/Velnbur/uniswapv2-indexer/internal/contracts"
)

func InitUniswapV2(
	logger *logan.Entry, client *ethclient.Client, redis *redis.Client,
	factoryAddr common.Address,
) (*contracts.UniswapV2, error) {
	logger.WithField("service", "info").Info("Initializing UniswapV2 contracts")
	ci, err := NewContractInitializer(
		logger, client, redis, factoryAddr,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create contract initializer")
	}

	return ci.Init(context.Background())
}
