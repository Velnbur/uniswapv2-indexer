package handlers

import (
	"net/http"

	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"

	"github.com/Velnbur/uniswapv2-indexer/internal/service/api/requests"
)

// Algorithm

func GetBestPath(w http.ResponseWriter, r *http.Request) {
	req, err := requests.NewBestPathRequest(r)
	if err != nil {
		Log(r).WithError(err).Debug("invalid request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	pathes, err := PathesProvider(r).GetPathes(r.Context(), req.TokenIn, req.TokenOut)
	if err != nil {
		Log(r).WithError(err).Error("failed to get pathes from provider")
		ape.RenderErr(w, problems.InternalError())
		return
	}
}
