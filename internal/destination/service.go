package destination

import (
	"context"
	"fmt"
)

type Repository interface {
	GetByCity(ctx context.Context, city string) (*Destination, error)
	Upsert(ctx context.Context, d *Destination) error
}

type Cache interface {
	Get(ctx context.Context, key string) (*Destination, error)
	Set(ctx context.Context, key string, d *Destination) error
	Delete(ctx context.Context, key string) error
}

type Fetcher interface {
	Fetch(ctx context.Context, city string) (*Destination, error)
}

type Service struct {
	repo    Repository
	cache   Cache
	fetcher Fetcher
}

func NewService(repo Repository, cache Cache, fetcher Fetcher) *Service {
	return &Service{repo: repo, cache: cache, fetcher: fetcher}
}

func (s *Service) GetByCity(ctx context.Context, city string) (*Destination, error) {
	key := cacheKey(city)

	if d, err := s.cache.Get(ctx, key); err == nil {
		return d, nil
	}

	d, err := s.repo.GetByCity(ctx, city)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, key, d)
	return d, nil
}

func (s *Service) Refresh(ctx context.Context, city string) (*Destination, error) {
	d, err := s.fetcher.Fetch(ctx, city)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", city, err)
	}

	if err := s.repo.Upsert(ctx, d); err != nil {
		return nil, fmt.Errorf("upsert %s: %w", city, err)
	}

	_ = s.cache.Set(ctx, cacheKey(city), d)
	return d, nil
}

func cacheKey(city string) string {
	return fmt.Sprintf("destination:%s", city)
}
