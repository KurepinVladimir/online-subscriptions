package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/KurepinVladimir/online-subscriptions/internal/model"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, s *model.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	Update(ctx context.Context, s *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter model.ListFilter) ([]model.Subscription, error)
	ListForPeriod(ctx context.Context, filter model.PeriodFilter) ([]model.Subscription, error)
}

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Start Migrations: all .sql from  migrations  a-z.
func RunMigrations(db *sqlx.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}
		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("exec migration %s: %w", f, err)
		}
	}

	return nil
}

func (r *PostgresRepository) Create(ctx context.Context, s *model.Subscription) error {
	query := `
INSERT INTO subscriptions (id, service_name, price, user_id, start_month, end_month)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING created_at, updated_at`
	return r.db.QueryRowContext(
		ctx,
		query,
		s.ID, s.ServiceName, s.Price, s.UserID, s.StartMonth, s.EndMonth,
	).Scan(&s.CreatedAt, &s.UpdatedAt)
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_month, end_month, created_at, updated_at
FROM subscriptions WHERE id = $1`

	var s model.Subscription
	if err := r.db.GetContext(ctx, &s, query, id); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepository) Update(ctx context.Context, s *model.Subscription) error {
	query := `
UPDATE subscriptions
SET service_name = $2,
    price        = $3,
    start_month  = $4,
    end_month    = $5,
    updated_at   = now()
WHERE id = $1
RETURNING created_at, updated_at`
	return r.db.QueryRowContext(
		ctx,
		query,
		s.ID, s.ServiceName, s.Price, s.StartMonth, s.EndMonth,
	).Scan(&s.CreatedAt, &s.UpdatedAt)
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	return err
}

func (r *PostgresRepository) List(ctx context.Context, filter model.ListFilter) ([]model.Subscription, error) {

	query := `
	SELECT id, service_name, price, user_id, start_month, end_month, created_at, updated_at
	FROM subscriptions
	WHERE 1=1`

	args := []any{}
	idx := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", idx)
		args = append(args, *filter.UserID)
		idx++
	}
	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", idx)
		args = append(args, *filter.ServiceName)
		idx++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", idx)
		args = append(args, filter.Limit)
		idx++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", idx)
		args = append(args, filter.Offset)
		idx++
	}

	var subs []model.Subscription
	if err := r.db.SelectContext(ctx, &subs, query, args...); err != nil {
		return nil, err
	}
	return subs, nil
}

// ListForPeriod returns subscriptions active during the specified period.
func (r *PostgresRepository) ListForPeriod(ctx context.Context, filter model.PeriodFilter) ([]model.Subscription, error) {

	query := `
	SELECT id, service_name, price, user_id, start_month, end_month, created_at, updated_at
	FROM subscriptions
	WHERE start_month <= $1
	AND (end_month IS NULL OR end_month >= $2)`
	// Important: $1 — period end, $2 — period start (intersection logic)

	args := []any{filter.To.Time, filter.From.Time}
	idx := 3

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", idx)
		args = append(args, *filter.UserID)
		idx++
	}
	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", idx)
		args = append(args, *filter.ServiceName)
		idx++
	}

	var subs []model.Subscription
	if err := r.db.SelectContext(ctx, &subs, query, args...); err != nil {
		return nil, err
	}
	return subs, nil
}
