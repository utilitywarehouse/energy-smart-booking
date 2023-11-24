package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/cache"
	"github.com/utilitywarehouse/go-operational/op"
)

const prefixKeyMeterpointEligibility = "mpe"

type MeterpointEligibleStore struct {
	r   *redis.Client
	ttl time.Duration
}

func NewMeterpointEligible(r *redis.Client, ttl time.Duration) *MeterpointEligibleStore {
	return &MeterpointEligibleStore{r: r, ttl: ttl}
}

func (s *MeterpointEligibleStore) NewHealthCheck() func(*op.CheckResponse) {
	return func(cr *op.CheckResponse) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.r.Info(ctx).Err(); err != nil {
			cr.Unhealthy("cant connect to redis: "+err.Error(), "Troubleshoot redis",
				"eligiblity cannot be cached")
			return
		}

		cr.Healthy("Redis is healthy")
	}
}

func (s *MeterpointEligibleStore) key(mpan, mprn string) string {
	if mprn == "" {
		return fmt.Sprintf("%s:%s", prefixKeyMeterpointEligibility, mpan)
	}
	return fmt.Sprintf("%s:%s:%s", prefixKeyMeterpointEligibility, mpan, mprn)
}

func (s *MeterpointEligibleStore) SetEligibilityForMpxn(ctx context.Context, mpan, mprn string, eligible bool) error {
	return s.r.Set(ctx, s.key(mpan, mprn), eligible, s.ttl).Err()
}

func (s *MeterpointEligibleStore) GetEligibilityForMpxn(ctx context.Context, mpan, mprn string) (bool, error) {
	e, err := s.r.Get(ctx, s.key(mpan, mprn)).Bool()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, cache.ErrNotFound
		}
		return false, err
	}
	return e, nil
}
