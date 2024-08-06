package types

import (
	"github.com/ethereum/go-ethereum/core/types"
)

type NewBlock struct {
	Header *types.Header
	Logs   []types.Log
}
