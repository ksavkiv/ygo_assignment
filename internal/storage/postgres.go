package storage

import (
	"context"
	"fmt"

	"destination-data-aggregation-api/internal/destination"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pool: pool}
}

func (r *PostgresRepo) GetAll(ctx context.Context) ([]destination.Destination, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, country, description, latitude, longitude, created_at, updated_at FROM destinations ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query destinations: %w", err)
	}
	defer rows.Close()

	var results []destination.Destination
	for rows.Next() {
		var d destination.Destination
		if err := rows.Scan(&d.ID, &d.Name, &d.Country, &d.Description, &d.Latitude, &d.Longitude, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan destination: %w", err)
		}
		results = append(results, d)
	}
	return results, rows.Err()
}

func (r *PostgresRepo) GetByID(ctx context.Context, id int64) (*destination.Destination, error) {
	var d destination.Destination
	err := r.pool.QueryRow(ctx, `SELECT id, name, country, description, latitude, longitude, created_at, updated_at FROM destinations WHERE id = $1`, id).
		Scan(&d.ID, &d.Name, &d.Country, &d.Description, &d.Latitude, &d.Longitude, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get destination %d: %w", id, err)
	}
	return &d, nil
}

func (r *PostgresRepo) Create(ctx context.Context, d *destination.Destination) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO destinations (name, country, description, latitude, longitude) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`,
		d.Name, d.Country, d.Description, d.Latitude, d.Longitude,
	).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	return nil
}

func (r *PostgresRepo) Update(ctx context.Context, d *destination.Destination) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE destinations SET name=$1, country=$2, description=$3, latitude=$4, longitude=$5, updated_at=NOW() WHERE id=$6`,
		d.Name, d.Country, d.Description, d.Latitude, d.Longitude, d.ID,
	)
	if err != nil {
		return fmt.Errorf("update destination %d: %w", d.ID, err)
	}
	return nil
}

func (r *PostgresRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM destinations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete destination %d: %w", id, err)
	}
	return nil
}
