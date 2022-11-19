package config

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type Ethereumer interface {
	EthereumCfg() EthereumCfg
	EthereumClient() *ethclient.Client
}

type EthereumCfg struct {
	Node string `fig:"node,required"`
}

func NewEthereumCfg(getter kv.Getter) Ethereumer {
	return &ethereumCfg{
		getter: getter,
	}
}

type ethereumCfg struct {
	getter kv.Getter
	once   comfig.Once
}

const yamlEthereumerKey = "ethereum"

func (c *ethereumCfg) EthereumCfg() EthereumCfg {
	return c.once.Do(func() interface{} {
		var cfg EthereumCfg

		err := figure.Out(&cfg).
			From(kv.MustGetStringMap(c.getter, yamlEthereumerKey)).
			Please()
		if err != nil {
			panic(err)
		}

		return cfg
	}).(EthereumCfg)
}

func (c *ethereumCfg) EthereumClient() *ethclient.Client {
	return c.once.Do(func() interface{} {
		client, err := ethclient.Dial(c.EthereumCfg().Node)
		if err != nil {
			panic(errors.Wrap(err, "failed to connect to ethereum node"))
		}

		return client
	}).(*ethclient.Client)
}
