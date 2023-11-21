package cache

import (
	"context"
	"errors"

	log "github.com/sirupsen/logrus"
)

var ErrNotFound = errors.New("not cached")

type EligibilityGateway interface {
	GetMeterpointEligibility(ctx context.Context, mpan, mprn, postcode string) (bool, error)
}

type EligibilityCache interface {
	GetEligibilityForMpxn(ctx context.Context, mpan, mprn string) (bool, error)
	SetEligibilityForMpxn(ctx context.Context, mpan, mprn string, eligible bool) error
}

type MeterpointEligibilityCacheWrapper struct {
	gw    EligibilityGateway
	cache EligibilityCache
}

func NewMeterpointEligibilityCacheWrapper(gw EligibilityGateway, cache EligibilityCache) *MeterpointEligibilityCacheWrapper {
	return &MeterpointEligibilityCacheWrapper{
		gw:    gw,
		cache: cache,
	}
}

func (c *MeterpointEligibilityCacheWrapper) GetMeterpointEligibility(ctx context.Context, mpan, mprn, postcode string) (bool, error) {
	eligible, err := c.cache.GetEligibilityForMpxn(ctx, mpan, mprn)
	if err == nil {
		return eligible, nil
	}

	if !errors.Is(err, ErrNotFound) {
		return false, err
	}

	eligible, err = c.gw.GetMeterpointEligibility(ctx, mpan, mprn, postcode)
	if err != nil {
		return false, err
	}

	err = c.cache.SetEligibilityForMpxn(ctx, mpan, mprn, eligible)
	if err != nil {
		log.WithField("mpan", mpan).Warn("unable to write to cache")
	}
	return eligible, nil
}
