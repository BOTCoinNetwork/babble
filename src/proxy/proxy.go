package proxy

import (
	"github.com/BOTCoinNetwork/babble/src/hashgraph"
)

// AppProxy ...
type AppProxy interface {
	SubmitCh() chan []byte
	CommitBlock(block hashgraph.Block) (CommitResponse, error)
	GetSnapshot(blockIndex int) ([]byte, error)
	Restore(snapshot []byte) error
}
