package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/Velnbur/uniswapv2-indexer/internal/data"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

var _ PathesProvider = &PathesRedisProvider{}

type PathesRedisProvider struct {
	cache *redis.Client
}

func NewPathesRedisProvider(client *redis.Client) *PathesRedisProvider {
	return &PathesRedisProvider{
		cache: client,
	}
}

const pathesKey = "pathes:%s-%s"

func (p *PathesRedisProvider) GetPathes(
	ctx context.Context, token0, token1 common.Address,
) ([]data.Path, error) {
	key := fmt.Sprintf(pathesKey, token0.Hex(), token1.Hex())

	raw, err := p.cache.Get(ctx, key).Result()

	pathes := p.parsePathes(raw)

	switch err {
	case nil:
		return pathes, nil
	case redis.Nil:
		return []data.Path{}, nil
	default:
		return []data.Path{}, errors.Wrap(err, "failed to get erc20 symbol")
	}
}

const (
	pathesSeparator      = ";"
	pathElementSeparator = ","
)

func (p *PathesRedisProvider) parsePathes(raw string) []data.Path {
	rawPathes := strings.Split(raw, pathesSeparator)

	pathes := make([]data.Path, len(rawPathes))

	for i, rawPath := range rawPathes {
		pathElements := strings.Split(rawPath, pathElementSeparator)

		pathes[i] = make([]common.Address, len(pathElements))

		for j, element := range pathElements {
			pathes[i][j] = common.HexToAddress(element)
		}
	}

	return pathes
}

func (p *PathesRedisProvider) SetPathes(
	ctx context.Context, token0, token1 common.Address, pathes []data.Path,
) error {
	key := fmt.Sprintf(pathesKey, token0, token1)

	raw := p.pathesToString(pathes)

	return p.cache.Set(ctx, key, raw, 0).Err()
}

func (p *PathesRedisProvider) pathesToString(pathes []data.Path) string {
	rawPathes := make([]string, len(pathes))

	for i, path := range pathes {
		rawPath := make([]string, len(path))

		for j, elem := range path {
			rawPath[j] = elem.Hex()
		}

		rawPathes[i] = strings.Join(rawPath, pathElementSeparator)
	}

	return strings.Join(rawPathes, pathesSeparator)
}
