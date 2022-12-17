package requests

import (
	"github.com/ethereum/go-ethereum/common"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func isHexAddress(value interface{}) error {
	address, ok := value.(*string)
	if !ok {
		return errors.New("invalid address type")
	}

	if ok = common.IsHexAddress(*address); !ok {
		return errors.New("not a valid hex address")
	}

	return nil
}
