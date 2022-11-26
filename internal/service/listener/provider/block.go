package provider

import "context"

type CurrentBlockProvider interface {
	CurrentBlock(ctx context.Context) (uint64, error)
	UpdateBlock(ctx context.Context, block uint64) error
}
