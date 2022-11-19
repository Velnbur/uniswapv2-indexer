package config

import (
	"github.com/Velnbur/uniswapv2-indexer/internal/contracts"
	"github.com/go-redis/redis/v8"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/copus"
	"gitlab.com/distributed_lab/kit/copus/types"
	"gitlab.com/distributed_lab/kit/kv"
)

type Config interface {
	comfig.Logger
	types.Copuser
	comfig.Listenerer

	Contracter
	Ethereumer

	Redis() *redis.Client
	UniswapV2() *contracts.UniswapV2
}

type config struct {
	comfig.Logger
	types.Copuser
	comfig.Listenerer
	getter kv.Getter

	Contracter
	Ethereumer

	redis     comfig.Once
	uniswapv2 comfig.Once
}

func New(getter kv.Getter) Config {
	return &config{
		getter:     getter,
		Copuser:    copus.NewCopuser(getter),
		Listenerer: comfig.NewListenerer(getter),
		Logger:     comfig.NewLogger(getter, comfig.LoggerOpts{}),
		Contracter: NewContracterCfg(getter),
		Ethereumer: NewEthereumCfg(getter),
	}
}
