package listener

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"

	"github.com/Velnbur/uniswapv2-indexer/internal/channels"
	"github.com/Velnbur/uniswapv2-indexer/internal/contracts"
)

func (l *Listener) Listen(ctx context.Context) error {
	query, err := l.filters(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to set subscription filters")
	}

	logs := make(chan types.Log)
	sub, err := l.client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to logs")
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sub.Err():
			return errors.Wrap(err, "failed to subscribe to logs")
		case vLog := <-logs:
			if err := l.proccessLog(ctx, &vLog); err != nil {
				l.logger.WithError(err).Error("failed to handle event")
			}
		}
	}
}

func (l *Listener) proccessLog(ctx context.Context, log *types.Log) error {
	block, err := l.currentBlock.CurrentBlock(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get current block")
	}
	if block < log.BlockNumber {
		block = log.BlockNumber
		if err := l.currentBlock.UpdateBlock(ctx, block); err != nil {
			return errors.Wrap(err, "failed to update current block")
		}

		err = l.eventQueue.Send(ctx, channels.Event{
			Type: channels.BlockCreationEvent,
			BlockCreation: &channels.BlockCreation{
				Block: block,
			},
		})
		if err != nil {
			return errors.Wrap(err, "failed to send block creation event")
		}
	}

	if err := l.handleEvent(ctx, log); err != nil {
		return errors.Wrap(err, "failed to update current block")
	}
	return nil
}

func (l *Listener) filters(ctx context.Context) (ethereum.FilterQuery, error) {
	addresses := l.getAllPairsAddresses()

	block, err := l.currentBlock.CurrentBlock(ctx)
	if err != nil {
		return ethereum.FilterQuery{}, errors.Wrap(err,
			"failed to get last block",
		)
	}

	var blockInt *big.Int = nil
	// that means, that we weren't initialized yet,
	// and we need to start indexing from last block
	if block != 0 {
		// other wise, start indexing from last saved block
		blockInt = new(big.Int).SetUint64(block)
	}

	topics := make([]common.Hash, 0)
	for _, event := range AllEvents() {
		topic, ok := l.pairABI.Events[string(event)]
		if !ok {
			return ethereum.FilterQuery{}, errors.Wrap(err,
				"no such event in abi",
				logan.F{
					"event": event,
				})
		}
		topics = append(topics, topic.ID)
	}
	topic, ok := l.factoryABI.Events["PairCreated"]
	if !ok {
		return ethereum.FilterQuery{}, errors.Wrap(err,
			"no such event in factory abi",
		)
	}
	topics = append(topics, topic.ID)

	query := ethereum.FilterQuery{
		FromBlock: blockInt,
		Addresses: addresses,
		Topics: [][]common.Hash{
			topics,
		},
	}

	return query, nil
}

func (l *Listener) getAllPairsAddresses() []common.Address {
	var addresses []common.Address

	l.uniswapV2.Pairs.Range(func(addr common.Address, _ *contracts.UniswapV2Pair) bool {
		addresses = append(addresses, addr)
		return true
	})

	return addresses
}
