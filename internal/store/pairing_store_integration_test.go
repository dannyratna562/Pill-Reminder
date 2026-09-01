package store

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/pillreminder/backend/internal/domain"
	"github.com/pillreminder/backend/internal/store/sqlcgen"
)

// These tests exercise PairingStore against a real Postgres database and
// need TEST_DATABASE_URL. They are not runnable in the sandbox this project
// was scaffolded in (no reachable Postgres) — run them once a real database
// is available, e.g.:
//
//	TEST_DATABASE_URL=postgres://user:pass@localhost:5432/pillreminder_test?sslmode=disable \
//	  go test ./internal/store/... -run Integration -v
func newIntegrationStore(t *testing.T) *PairingStore {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	pool, err := NewPool(context.Background(), dsn)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	t.Cleanup(pool.Close)
	return NewPairingStore(pool)
}

func TestPairingStore_RedeemAndLink_Integration(t *testing.T) {
	store := newIntegrationStore(t)
	ctx := context.Background()
	parentID := uuid.New()
	childID := uuid.New()

	code, err := store.CreatePairingCode(ctx, parentID)
	if err != nil {
		t.Fatalf("CreatePairingCode() error = %v", err)
	}

	link, err := store.RedeemAndLink(ctx, code.Code, childID)
	if err != nil {
		t.Fatalf("RedeemAndLink() error = %v", err)
	}
	if link.ParentID != parentID || link.ChildID != childID {
		t.Errorf("RedeemAndLink() = %+v, want ParentID=%v ChildID=%v", link, parentID, childID)
	}
}

func TestPairingStore_RedeemAndLink_ConcurrentDoubleRedeem_Integration(t *testing.T) {
	store := newIntegrationStore(t)
	ctx := context.Background()
	parentID := uuid.New()

	code, err := store.CreatePairingCode(ctx, parentID)
	if err != nil {
		t.Fatalf("CreatePairingCode() error = %v", err)
	}

	const racers = 10
	var wg sync.WaitGroup
	var successes int
	var mu sync.Mutex
	wg.Add(racers)
	for i := 0; i < racers; i++ {
		go func() {
			defer wg.Done()
			if _, err := store.RedeemAndLink(ctx, code.Code, uuid.New()); err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if successes != 1 {
		t.Errorf("concurrent RedeemAndLink() succeeded %d times, want exactly 1", successes)
	}
}

func TestPairingStore_RedeemAndLink_Expired_Integration(t *testing.T) {
	store := newIntegrationStore(t)
	ctx := context.Background()

	// Directly insert an already-expired code via the store's own queries so
	// we don't need a second code path just for this test.
	row, err := store.queries.CreatePairingCode(ctx, createExpiredParams(uuid.New()))
	if err != nil {
		t.Fatalf("CreatePairingCode() error = %v", err)
	}

	_, err = store.RedeemAndLink(ctx, row.Code, uuid.New())
	if !errors.Is(err, domain.ErrPairingCodeExpired) {
		t.Errorf("RedeemAndLink() error = %v, want ErrPairingCodeExpired", err)
	}
}

func TestPairingStore_RedeemAndLink_AlreadyUsed_Integration(t *testing.T) {
	store := newIntegrationStore(t)
	ctx := context.Background()

	code, err := store.CreatePairingCode(ctx, uuid.New())
	if err != nil {
		t.Fatalf("CreatePairingCode() error = %v", err)
	}
	if _, err := store.RedeemAndLink(ctx, code.Code, uuid.New()); err != nil {
		t.Fatalf("first RedeemAndLink() error = %v", err)
	}

	_, err = store.RedeemAndLink(ctx, code.Code, uuid.New())
	if !errors.Is(err, domain.ErrPairingCodeUsed) {
		t.Errorf("second RedeemAndLink() error = %v, want ErrPairingCodeUsed", err)
	}
}

func createExpiredParams(parentID uuid.UUID) sqlcgen.CreatePairingCodeParams {
	code, err := domain.GenerateCode()
	if err != nil {
		panic(err)
	}
	return sqlcgen.CreatePairingCodeParams{
		Code:      code,
		ParentID:  parentID,
		ExpiresAt: time.Now().Add(-time.Minute),
	}
}
