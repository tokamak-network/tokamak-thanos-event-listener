package listener

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/bcclient"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/constant"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/repository"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/testutil"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/types"
	"golang.org/x/sync/errgroup"
)

const (
	rpcUrl = "ws://sepolia.rpc.tokamak.network:8546"
)

func Test_syncOldBlocks(t *testing.T) {
	const (
		totalOldBlocks = uint64(5)
	)
	ctx := context.Background()

	bcClient, err := bcclient.New(ctx, rpcUrl)
	require.NoError(t, err)

	syncBlockKeeper := &testutil.SyncBlockInMemKeeper{}
	keeper, err := repository.NewBlockKeeper(ctx, bcClient, syncBlockKeeper)
	require.NoError(t, err)

	listenerSrv, err := MakeService("test-event-listener", bcClient, keeper)
	require.NoError(t, err)

	currentHead, err := bcClient.BlockNumber(ctx)
	require.NoError(t, err)

	oldBlockNumber := currentHead - totalOldBlocks
	consumingBlock, err := bcClient.HeaderAtBlockNumber(ctx, oldBlockNumber)
	require.NoError(t, err)

	err = keeper.SetHead(ctx, consumingBlock, constant.ZeroHash)
	require.NoError(t, err)
	oldBlocksCh := make(chan *types.NewBlock)

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		err = listenerSrv.syncOldBlocks(ctx, oldBlocksCh)
		defer close(oldBlocksCh)
		return err
	})

	i := uint64(0)
	blockNo := oldBlockNumber + 1
	for block := range oldBlocksCh {
		currentBlockNumber := block.Header.Number.Uint64()
		assert.Equal(t, currentBlockNumber, blockNo)

		t.Logf(`Got old block: %d`, block.Header.Number.Uint64())
		i++
		blockNo++
	}
	assert.Equal(t, i, totalOldBlocks)

	require.NoError(t, g.Wait())
}

func Test_handleReorgBlock(t *testing.T) {
	ctx := context.Background()

	bcClient, err := bcclient.New(ctx, rpcUrl)
	require.NoError(t, err)

	syncBlockKeeper := &testutil.SyncBlockInMemKeeper{}
	keeper, err := repository.NewBlockKeeper(ctx, bcClient, syncBlockKeeper)
	require.NoError(t, err)

	listenerSrv, err := MakeService("test-event-listener", bcClient, keeper)
	require.NoError(t, err)
	header, err := bcClient.GetHeader(ctx)
	require.NoError(t, err)

	err = keeper.SetHead(ctx, header, constant.ZeroHash)
	require.NoError(t, err)

	// This causes the gap between the current head in the keeper and the latest head at least two blocks
	time.Sleep(24 * time.Second)

	header, err = bcClient.GetHeader(ctx)
	require.NoError(t, err)

	blocks, err := listenerSrv.handleReorgBlocks(ctx, header)
	require.NoError(t, err)

	assert.Equal(t, true, len(blocks) > 0)

}
