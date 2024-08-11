package erc20

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/bcclient"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/types"
)

func Test_FetchTokenInfo(t *testing.T) {
	ctx := context.Background()

	type testCases []struct {
		Expected        types.Token
		ContractAddress string
	}

	bcClient, err := bcclient.New(ctx, "https://sepolia.rpc.tokamak.network")
	require.NoError(t, err)

	var tests = testCases{
		{
			Expected: types.Token{
				Symbol:   "TON",
				Decimals: 18,
				Address:  strings.ToLower("0xa30fe40285B8f5c0457DbC3B7C8A280373c40044"),
			},
			ContractAddress: strings.ToLower("0xa30fe40285B8f5c0457DbC3B7C8A280373c40044"),
		},
		{
			Expected: types.Token{
				Symbol:   "TOS",
				Decimals: 18,
				Address:  strings.ToLower("0xFF3Ef745D9878AfE5934Ff0b130868AFDDbc58e8"),
			},
			ContractAddress: strings.ToLower("0xFF3Ef745D9878AfE5934Ff0b130868AFDDbc58e8"),
		},
	}
	t.Parallel()
	for _, test := range tests {
		t.Run(test.ContractAddress, func(t *testing.T) {
			tokenInfo, err := FetchTokenInfo(bcClient, test.ContractAddress)
			require.NoError(t, err)

			assert.NotEmpty(t, tokenInfo)
			assert.Equal(t, test.Expected.Symbol, tokenInfo.Symbol)
			assert.Equal(t, test.Expected.Decimals, tokenInfo.Decimals)
			assert.Equal(t, test.ContractAddress, tokenInfo.Address)
		})
	}

}
