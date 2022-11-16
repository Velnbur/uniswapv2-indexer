package config

import (
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type Ethereumer interface {
	EthereumCfg() EthereumCfg
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
