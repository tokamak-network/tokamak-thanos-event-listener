package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/bcclient"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/testutil"
)

var (
	wsUrl   = "ws://sepolia.rpc.tokamak.network:8546"
	httpUrl = "https://sepolia.rpc.tokamak.network"
)

func TestBlockKeeper_initWithExistingBlockHash(t *testing.T) {
	ctx := context.Background()

	syncBlockKeeper := &testutil.SyncBlockInMemKeeper{}

	bcClient, err := bcclient.New(ctx, wsUrl, httpUrl)
	require.NoError(t, err)

	blockNo, err := bcClient.BlockNumber(ctx)
	require.NoError(t, err)

	block, err := bcClient.HeaderAtBlockNumber(ctx, blockNo)
	require.NoError(t, err)

	// set the consuming block hash
	err = syncBlockKeeper.SetHead(ctx, block.Hash().String())
	require.NoError(t, err)

	blockKeeper, err := NewBlockKeeper(ctx, bcClient, syncBlockKeeper)
	require.NoError(t, err)

	assert.Equal(t, TwoEpochBlocks, blockKeeper.q.Size())
	assert.Equal(t, TwoEpochBlocks, len(blockKeeper.blocks))
	assert.Equal(t, block.Hash(), blockKeeper.head.Hash())
}

func TestBlockKeeper_initWithoutExistingBlockHash(t *testing.T) {
	ctx := context.Background()

	syncBlockKeeper := &testutil.SyncBlockInMemKeeper{}

	bcClient, err := bcclient.New(ctx, wsUrl, httpUrl)
	require.NoError(t, err)

	currentBlock, err := bcClient.GetHeader(ctx)
	require.NoError(t, err)

	blockKeeper, err := NewBlockKeeper(ctx, bcClient, syncBlockKeeper)
	require.NoError(t, err)

	assert.Equal(t, TwoEpochBlocks, blockKeeper.q.Size())
	assert.Equal(t, TwoEpochBlocks, len(blockKeeper.blocks))
	assert.Equal(t, currentBlock.Hash(), blockKeeper.head.Hash())
}

func TestBlockKeeper_getReorgBlocks(t *testing.T) {
	ctx := context.Background()

	syncBlockKeeper := &testutil.SyncBlockInMemKeeper{}

	bcClient, err := bcclient.New(ctx, wsUrl, httpUrl)
	require.NoError(t, err)

	blockNo, err := bcClient.BlockNumber(ctx)
	require.NoError(t, err)

	block, err := bcClient.HeaderAtBlockNumber(ctx, blockNo-5)
	require.NoError(t, err)

	currentBlock, err := bcClient.GetHeader(ctx)
	require.NoError(t, err)

	// set the consuming block hash
	err = syncBlockKeeper.SetHead(ctx, block.Hash().String())
	require.NoError(t, err)

	blockKeeper, err := NewBlockKeeper(ctx, bcClient, syncBlockKeeper)
	require.NoError(t, err)

	assert.Equal(t, TwoEpochBlocks, blockKeeper.q.Size())
	assert.Equal(t, TwoEpochBlocks, len(blockKeeper.blocks))

	reorgedBlocks, _, err := blockKeeper.GetReorgHeaders(ctx, currentBlock)
	require.NoError(t, err)

	assert.Equal(t, true, len(reorgedBlocks) >= 4)
	assert.Equal(t, true, len(reorgedBlocks) <= 5)
}
