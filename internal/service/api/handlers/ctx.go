package handlers

import (
	"context"
	"net/http"

	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
	"gitlab.com/distributed_lab/logan/v3"
)

type ctxKey int

const (
	logCtxKey ctxKey = iota
	pathesProviderKey
)

func CtxLog(entry *logan.Entry) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, logCtxKey, entry)
	}
}

func Log(r *http.Request) *logan.Entry {
	return r.Context().Value(logCtxKey).(*logan.Entry)
}

func CtxPathesProvider(entry providers.PathesProvider) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, pathesProviderKey, entry)
	}
}

func PathesProvider(r *http.Request) providers.PathesProvider {
	return r.Context().Value(pathesProviderKey).(providers.PathesProvider)
}
