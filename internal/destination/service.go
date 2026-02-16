package destination

import (
	"context"
	"fmt"
)

type Repository interface {
	GetAll(ctx context.Context) ([]Destination, error)
	GetByID(ctx context.Context, id int64) (*Destination, error)
	Create(ctx context.Context, d *Destination) error
	Update(ctx context.Context, d *Destination) error
	Delete(ctx context.Context, id int64) error
}

type Cache interface {
	Get(ctx context.Context, key string) (*Destination, error)
	Set(ctx context.Context, key string, d *Destination) error
	Delete(ctx context.Context, key string) error
}

type Service struct {
	repo  Repository
	cache Cache
}

func NewService(repo Repository, cache Cache) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) List(ctx context.Context) ([]Destination, error) {
	return s.repo.GetAll(ctx)
}

func (s *Service) Get(ctx context.Context, id int64) (*Destination, error) {
	key := cacheKey(id)

	if d, err := s.cache.Get(ctx, key); err == nil {
		return d, nil
	}

	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, key, d)
	return d, nil
}

func (s *Service) Create(ctx context.Context, d *Destination) error {
	return s.repo.Create(ctx, d)
}

func (s *Service) Update(ctx context.Context, d *Destination) error {
	if err := s.repo.Update(ctx, d); err != nil {
		return err
	}
	_ = s.cache.Delete(ctx, cacheKey(d.ID))
	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.cache.Delete(ctx, cacheKey(id))
	return nil
}

func cacheKey(id int64) string {
	return fmt.Sprintf("destination:%d", id)
}
