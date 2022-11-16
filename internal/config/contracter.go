package config

import (
	"github.com/ethereum/go-ethereum/common"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type Contracter interface {
	ContracterCfg() ContracterCfg
}

type ContracterCfg struct {
	Factory common.Address
}

func NewContracterCfg(getter kv.Getter) Contracter {
	return &contracter{
		getter: getter,
	}
}

type contracter struct {
	getter kv.Getter
	once   comfig.Once
}

type contracterCfg struct {
	Factory string `fig:"factory,required"`
}

const yamlContracterKey = "contracts"

func (c *contracter) ContracterCfg() ContracterCfg {
	return c.once.Do(func() interface{} {
		var cfg contracterCfg

		err := figure.Out(&cfg).
			From(kv.MustGetStringMap(c.getter, yamlContracterKey)).
			Please()
		if err != nil {
			panic(err)
		}

		return ContracterCfg{
			Factory: common.HexToAddress(cfg.Factory),
		}
	}).(ContracterCfg)
}
