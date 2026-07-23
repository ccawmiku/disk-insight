package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/ccawmiku/disk-insight/internal/model"
)

func TestAllocatedSizeForRunDeduplicatesHardLinks(t *testing.T) {
	dataStore, err := Open(filepath.Join(t.TempDir(), "store.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = dataStore.Close() })
	ctx := context.Background()
	if err := dataStore.SyncRoots(ctx, []model.RootConfig{{Name: "Test", Path: t.TempDir(), Enabled: true}}); err != nil {
		t.Fatal(err)
	}
	roots, err := dataStore.RootConfigs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	runID, _, err := dataStore.StartRun(ctx, roots[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	hardLinkBytes := int64(4096)
	ordinaryBytes := int64(2048)
	now := time.Now()
	if err := dataStore.InsertEntries(ctx, roots[0].ID, runID, []Entry{
		{Path: "a", Name: "a", Kind: "file", AllocatedSize: &hardLinkBytes, Identity: "1:2", ModifiedAt: now},
		{Path: "b", Name: "b", Kind: "file", AllocatedSize: &hardLinkBytes, Identity: "1:2", ModifiedAt: now},
		{Path: "c", Name: "c", Kind: "file", AllocatedSize: &ordinaryBytes, ModifiedAt: now},
	}); err != nil {
		t.Fatal(err)
	}
	got, err := dataStore.AllocatedSizeForRun(ctx, roots[0].ID, runID)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || *got != hardLinkBytes+ordinaryBytes {
		t.Fatalf("allocated size = %v, want %d", got, hardLinkBytes+ordinaryBytes)
	}
}
