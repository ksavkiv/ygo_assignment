package storage

import (
	"context"
	"fmt"
	"log/slog"

	"destination-data-aggregation-api/internal/destination"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pool: pool}
}

func (r *PostgresRepo) GetByCity(ctx context.Context, city string) (*destination.Destination, error) {
	var d destination.Destination
	err := r.pool.QueryRow(ctx,
		`SELECT id, city, country, latitude, longitude, metadata, created_at, updated_at,
		        metadata->'weather'->>'current_temp_c' AS current_temp,
		        metadata->'country'->>'region' AS region,
		        metadata->'safety'->>'score' AS safety_score
		 FROM destinations WHERE city = $1`, city).
		Scan(&d.ID, &d.City, &d.Country, &d.Latitude, &d.Longitude, &d.Metadata, &d.CreatedAt, &d.UpdatedAt,
			&d.CurrentTemp, &d.Region, &d.SafetyScore)
	if err != nil {
		slog.Error("postgres: failed to get destination", "city", city, "error", err)
		return nil, fmt.Errorf("get destination %s: %w", city, err)
	}
	return &d, nil
}


func (r *PostgresRepo) ListCities(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT city FROM destinations ORDER BY city`)
	if err != nil {
		slog.Error("postgres: failed to list cities", "error", err)
		return nil, fmt.Errorf("list cities: %w", err)
	}
	defer rows.Close()

	var cities []string
	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err != nil {
			return nil, fmt.Errorf("scan city: %w", err)
		}
		cities = append(cities, city)
	}
	return cities, rows.Err()
}

func (r *PostgresRepo) Upsert(ctx context.Context, d *destination.Destination) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO destinations (city, country, latitude, longitude, metadata)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (city) DO UPDATE SET
		   country = EXCLUDED.country,
		   latitude = EXCLUDED.latitude,
		   longitude = EXCLUDED.longitude,
		   metadata = EXCLUDED.metadata,
		   updated_at = NOW()
		 RETURNING id, created_at, updated_at`,
		d.City, d.Country, d.Latitude, d.Longitude, d.Metadata,
	).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		slog.Error("postgres: failed to upsert destination", "city", d.City, "error", err)
		return fmt.Errorf("upsert destination %s: %w", d.City, err)
	}
	return nil
}
