package requests

import (
	"math/big"
	"net/http"

	"github.com/Velnbur/uniswapv2-indexer/pkg/helpers"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/urlval"
)

type bestPathRequestUrlParams struct {
	TokenIn  string `url:"token_in"`
	TokenOut string `url:"token_out"`
	AmountIn string `url:"amount_in"`
}

func (p bestPathRequestUrlParams) Validate() error {
	err := validation.Errors{
		"token_in":  validation.Validate(&p.TokenIn, validation.By(isHexAddress)),
		"token_out": validation.Validate(&p.TokenOut, validation.By(isHexAddress)),
		"amount_in": validation.Validate(&p.AmountIn, validation.Length(0, 100)), // TODO:
	}

	return err.Filter()
}

type BestPathRequest struct {
	TokenIn  common.Address
	TokenOut common.Address
	AmountIn *big.Int
}

func (req BestPathRequest) Validate() error {
	errs := validation.Errors{
		"token_in":  validation.Validate(&req.TokenIn, validation.NotIn(helpers.ZeroAddress)),
		"token_out": validation.Validate(&req.TokenOut, validation.NotIn(helpers.ZeroAddress)),
		"amount_in": validation.Validate(&req.AmountIn,
			validation.Min(big.NewInt(1)),
			validation.Max(math.MaxBig256),
		),
	}

	return errs.Filter()
}

func NewBestPathRequest(r *http.Request) (*BestPathRequest, error) {
	var params bestPathRequestUrlParams

	if err := urlval.Decode(r.URL.Query(), &params); err != nil {
		return nil, errors.Wrap(err, "invalid parameters in url for encoding")
	}

	if err := params.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid parameters in url")
	}

	amountIn, ok := new(big.Int).SetString(params.AmountIn, 10)
	if !ok {
		return nil, errors.New("invalid amount_in")
	}

	req := &BestPathRequest{
		TokenIn:  common.HexToAddress(params.TokenIn),
		TokenOut: common.HexToAddress(params.TokenOut),
		AmountIn: amountIn,
	}

	return req, req.Validate()
}
