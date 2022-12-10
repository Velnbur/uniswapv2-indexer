package listener

type EventKey string

const (
	SwapEvent EventKey = "Swap"
	SyncEvent EventKey = "Sync"
	MintEvent EventKey = "Mint"
	BurnEvent EventKey = "Burn"
)

func AllEvents() []EventKey {
	return []EventKey{
		SwapEvent,
		SyncEvent,
		BurnEvent,
		MintEvent,
	}
}
