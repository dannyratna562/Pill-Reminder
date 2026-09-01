package store

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pillreminder/backend/internal/domain"
	"github.com/pillreminder/backend/internal/store/sqlcgen"
)

// This file exercises the pill_schedules table schema itself (constraints,
// indexes, defaults) directly against a real Postgres database via
// TEST_DATABASE_URL. It uses inline SQL rather than sqlc-generated queries
// since story-02 ships no query files for pill_schedules (those come in
// stories 03-06).
//
//	TEST_DATABASE_URL=postgres://user:pass@localhost:5432/pillreminder_test?sslmode=disable \
//	  go test ./internal/store/... -run PillSchedulesSchema -v
func newIntegrationPool(t *testing.T) *pgxpool.Pool {
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
	return pool
}

func TestPillSchedulesSchema_InsertAndReadBack_Integration(t *testing.T) {
	pool := newIntegrationPool(t)
	ctx := context.Background()
	parentID := uuid.New()

	row := pool.QueryRow(ctx, `
		INSERT INTO pill_schedules (parent_id, pill_name, times)
		VALUES ($1, 'Aspirin', ARRAY['08:00','20:30']::time[])
		RETURNING id, parent_id, pill_name, times, active, created_at, updated_at
	`, parentID)

	sched, err := scanPillSchedule(row)
	if err != nil {
		t.Fatalf("insert/scan error = %v", err)
	}

	if !sched.Active {
		t.Error("Active = false, want true (default)")
	}
	if sched.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero, want populated")
	}
	if sched.UpdatedAt.IsZero() {
		t.Error("UpdatedAt is zero, want populated")
	}
	if len(sched.Times) != 2 || sched.Times[0].String() != "08:00" || sched.Times[1].String() != "20:30" {
		t.Errorf("Times = %v, want [08:00 20:30]", sched.Times)
	}
	if sched.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", sched.ParentID, parentID)
	}
}

func TestPillSchedulesSchema_ChecksReject_Integration(t *testing.T) {
	pool := newIntegrationPool(t)
	ctx := context.Background()
	parentID := uuid.New()

	tests := []struct {
		name           string
		sql            string
		wantConstraint string
	}{
		{
			name:           "blank pill name",
			sql:            `INSERT INTO pill_schedules (parent_id, pill_name, times) VALUES ($1, '   ', ARRAY['08:00']::time[])`,
			wantConstraint: "pill_schedules_pill_name_not_blank",
		},
		{
			name:           "pill name over 100 chars",
			sql:            `INSERT INTO pill_schedules (parent_id, pill_name, times) VALUES ($1, repeat('a', 101), ARRAY['08:00']::time[])`,
			wantConstraint: "pill_schedules_pill_name_len",
		},
		{
			name:           "empty times array",
			sql:            `INSERT INTO pill_schedules (parent_id, pill_name, times) VALUES ($1, 'Aspirin', ARRAY[]::time[])`,
			wantConstraint: "pill_schedules_times_cardinality",
		},
		{
			name:           "null element in times array",
			sql:            `INSERT INTO pill_schedules (parent_id, pill_name, times) VALUES ($1, 'Aspirin', ARRAY['08:00', NULL]::time[])`,
			wantConstraint: "pill_schedules_times_no_nulls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pool.Exec(ctx, tt.sql, parentID)
			if err == nil {
				t.Fatalf("Exec() error = nil, want constraint violation for %s", tt.wantConstraint)
			}
			var pgErr *pgconn.PgError
			if !isPgError(err, &pgErr) {
				t.Fatalf("Exec() error = %v, want *pgconn.PgError", err)
			}
			if pgErr.ConstraintName != tt.wantConstraint {
				t.Errorf("ConstraintName = %q, want %q", pgErr.ConstraintName, tt.wantConstraint)
			}
		})
	}
}

func TestPillSchedulesSchema_PartialIndexExists_Integration(t *testing.T) {
	pool := newIntegrationPool(t)
	ctx := context.Background()

	var count int
	err := pool.QueryRow(ctx, `
		SELECT count(*) FROM pg_indexes
		WHERE tablename = 'pill_schedules' AND indexname = 'idx_pill_schedules_parent_active'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("QueryRow() error = %v", err)
	}
	if count != 1 {
		t.Errorf("idx_pill_schedules_parent_active count = %d, want 1", count)
	}
}

// scanPillSchedule scans a row (id, parent_id, pill_name, times, active,
// created_at, updated_at) and converts it into the domain representation
// via toDomainPillSchedule, exercising the same conversion path the store
// layer will use once stories 03-06 add real queries.
func scanPillSchedule(row pgx.Row) (domain.PillSchedule, error) {
	var (
		id        uuid.UUID
		parentID  uuid.UUID
		pillName  string
		times     []pgtype.Time
		active    bool
		createdAt time.Time
		updatedAt time.Time
	)
	if err := row.Scan(&id, &parentID, &pillName, &times, &active, &createdAt, &updatedAt); err != nil {
		return domain.PillSchedule{}, err
	}

	return toDomainPillSchedule(sqlcgen.PillSchedule{
		ID:        id,
		ParentID:  parentID,
		PillName:  pillName,
		Times:     times,
		Active:    active,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	})
}

// isPgError reports whether err wraps a *pgconn.PgError, storing it in target.
func isPgError(err error, target **pgconn.PgError) bool {
	return errors.As(err, target)
}
