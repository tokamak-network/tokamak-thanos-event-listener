package testutil

import "context"

type SyncBlockInMemKeeper struct {
	head string
}

func (k *SyncBlockInMemKeeper) GetHead(ctx context.Context) (string, error) {
	return k.head, nil
}

func (k *SyncBlockInMemKeeper) SetHead(ctx context.Context, head string) error {
	k.head = head
	return nil
}
