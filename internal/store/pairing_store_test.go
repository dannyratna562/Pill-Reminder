package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/pillreminder/backend/internal/store/sqlcgen"
)

// fakeQuerier implements sqlcgen.Querier for store-layer unit tests that
// don't need a real database.
type fakeQuerier struct {
	createPairingCodeFn func(ctx context.Context, arg sqlcgen.CreatePairingCodeParams) (sqlcgen.PairingCode, error)
	getPairingCodeFn    func(ctx context.Context, code string) (sqlcgen.PairingCode, error)
	redeemPairingCodeFn func(ctx context.Context, code string) (sqlcgen.PairingCode, error)
	createFamilyLinkFn  func(ctx context.Context, arg sqlcgen.CreateFamilyLinkParams) (sqlcgen.FamilyLink, error)
}

func (f *fakeQuerier) CreatePairingCode(ctx context.Context, arg sqlcgen.CreatePairingCodeParams) (sqlcgen.PairingCode, error) {
	return f.createPairingCodeFn(ctx, arg)
}

func (f *fakeQuerier) GetPairingCodeByCode(ctx context.Context, code string) (sqlcgen.PairingCode, error) {
	return f.getPairingCodeFn(ctx, code)
}

func (f *fakeQuerier) RedeemPairingCode(ctx context.Context, code string) (sqlcgen.PairingCode, error) {
	return f.redeemPairingCodeFn(ctx, code)
}

func (f *fakeQuerier) CreateFamilyLink(ctx context.Context, arg sqlcgen.CreateFamilyLinkParams) (sqlcgen.FamilyLink, error) {
	return f.createFamilyLinkFn(ctx, arg)
}

var _ sqlcgen.Querier = (*fakeQuerier)(nil)

func TestPairingStore_CreatePairingCode_RetriesOnCollision(t *testing.T) {
	parentID := uuid.New()
	calls := 0
	fake := &fakeQuerier{
		createPairingCodeFn: func(ctx context.Context, arg sqlcgen.CreatePairingCodeParams) (sqlcgen.PairingCode, error) {
			calls++
			if calls == 1 {
				return sqlcgen.PairingCode{}, &pgconn.PgError{Code: pgUniqueViolation}
			}
			return sqlcgen.PairingCode{
				ID:        uuid.New(),
				Code:      arg.Code,
				ParentID:  arg.ParentID,
				ExpiresAt: arg.ExpiresAt,
			}, nil
		},
	}
	store := &PairingStore{queries: fake}

	got, err := store.CreatePairingCode(context.Background(), parentID)
	if err != nil {
		t.Fatalf("CreatePairingCode() error = %v", err)
	}
	if calls != 2 {
		t.Errorf("CreatePairingCode() made %d attempts, want 2", calls)
	}
	if got.ParentID != parentID {
		t.Errorf("CreatePairingCode() ParentID = %v, want %v", got.ParentID, parentID)
	}
}

func TestPairingStore_CreatePairingCode_NonCollisionErrorStopsImmediately(t *testing.T) {
	calls := 0
	wantErr := context.DeadlineExceeded
	fake := &fakeQuerier{
		createPairingCodeFn: func(ctx context.Context, arg sqlcgen.CreatePairingCodeParams) (sqlcgen.PairingCode, error) {
			calls++
			return sqlcgen.PairingCode{}, wantErr
		},
	}
	store := &PairingStore{queries: fake}

	_, err := store.CreatePairingCode(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("CreatePairingCode() error = nil, want non-nil")
	}
	if calls != 1 {
		t.Errorf("CreatePairingCode() made %d attempts, want 1 (should not retry non-collision errors)", calls)
	}
}
