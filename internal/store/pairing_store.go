package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pillreminder/backend/internal/domain"
	"github.com/pillreminder/backend/internal/store/sqlcgen"
)

// maxCreateAttempts bounds the retry loop for CreatePairingCode's rare
// collision with another still-active code sharing the same digits.
const maxCreateAttempts = 5

// pgUniqueViolation is the Postgres error code for a unique constraint conflict.
const pgUniqueViolation = "23505"

// PairingStore persists pairing codes and the family_links they redeem into.
type PairingStore struct {
	pool    *pgxpool.Pool
	queries sqlcgen.Querier
}

func NewPairingStore(pool *pgxpool.Pool) *PairingStore {
	return &PairingStore{pool: pool, queries: sqlcgen.New(pool)}
}

// CreatePairingCode generates a fresh code and stores it, retrying on the
// rare collision with another active code sharing the same digits.
func (s *PairingStore) CreatePairingCode(ctx context.Context, parentID uuid.UUID) (domain.PairingCode, error) {
	var lastErr error
	for attempt := 0; attempt < maxCreateAttempts; attempt++ {
		code, err := domain.GenerateCode()
		if err != nil {
			return domain.PairingCode{}, fmt.Errorf("create pairing code: %w", err)
		}

		row, err := s.queries.CreatePairingCode(ctx, sqlcgen.CreatePairingCodeParams{
			Code:      code,
			ParentID:  parentID,
			ExpiresAt: time.Now().Add(domain.PairingCodeTTL),
		})
		if err == nil {
			return toDomainPairingCode(row), nil
		}
		if !isUniqueViolation(err) {
			return domain.PairingCode{}, fmt.Errorf("create pairing code: %w", err)
		}
		lastErr = err
	}
	return domain.PairingCode{}, fmt.Errorf("create pairing code: exhausted retries: %w", lastErr)
}

// RedeemAndLink atomically claims code (if it's unused and unexpired) and
// creates the resulting family_link, in one transaction.
func (s *PairingStore) RedeemAndLink(ctx context.Context, code string, childID uuid.UUID) (domain.FamilyLink, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.FamilyLink{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := sqlcgen.New(tx)

	claimed, err := q.RedeemPairingCode(ctx, code)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return domain.FamilyLink{}, fmt.Errorf("redeem pairing code: %w", err)
		}

		existing, getErr := q.GetPairingCodeByCode(ctx, code)
		if getErr != nil {
			if errors.Is(getErr, pgx.ErrNoRows) {
				return domain.FamilyLink{}, domain.ErrPairingCodeNotFound
			}
			return domain.FamilyLink{}, fmt.Errorf("get pairing code: %w", getErr)
		}

		pc := toDomainPairingCode(existing)
		if pc.IsExpired(time.Now()) {
			return domain.FamilyLink{}, domain.ErrPairingCodeExpired
		}
		// Either genuinely already used, or a concurrent redeemer won the
		// race between the failed claim above and this diagnostic read.
		return domain.FamilyLink{}, domain.ErrPairingCodeUsed
	}

	link, err := q.CreateFamilyLink(ctx, sqlcgen.CreateFamilyLinkParams{
		ChildID:  childID,
		ParentID: claimed.ParentID,
	})
	if err != nil {
		return domain.FamilyLink{}, fmt.Errorf("create family link: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.FamilyLink{}, fmt.Errorf("commit tx: %w", err)
	}
	return toDomainFamilyLink(link), nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}

func toDomainPairingCode(row sqlcgen.PairingCode) domain.PairingCode {
	return domain.PairingCode{
		ID:        row.ID,
		Code:      row.Code,
		ParentID:  row.ParentID,
		ExpiresAt: row.ExpiresAt,
		UsedAt:    timestamptzToPtr(row.UsedAt),
		CreatedAt: row.CreatedAt,
	}
}

func toDomainFamilyLink(row sqlcgen.FamilyLink) domain.FamilyLink {
	return domain.FamilyLink{
		ID:        row.ID,
		ChildID:   row.ChildID,
		ParentID:  row.ParentID,
		CreatedAt: row.CreatedAt,
	}
}

func timestamptzToPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	tt := t.Time
	return &tt
}
