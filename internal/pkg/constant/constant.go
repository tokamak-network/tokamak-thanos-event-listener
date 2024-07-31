package constant

import (
	"strings"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/types"
)

var (
	KnownTokensMap = map[string]types.Token{
		strings.ToLower("0x4200000000000000000000000000000000000486"): {
			Symbol:   "ETH",
			Decimals: 18,
		},
		strings.ToLower("0xDeadDeAddeAddEAddeadDEaDDEAdDeaDDeAD0000"): {
			Symbol:   "TON",
			Decimals: 18,
		},
	}
)
