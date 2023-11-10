package cache

import "context"

type EligibilityGateway interface {
	GetMeterpointEligibility(ctx context.Context, mpan, mprn, postcode string) (bool, error)
}

type EligibilityCache interface {
	GetEligibilityForMpxn(ctx context.Context, mpxn string) (bool, bool, error)
	CacheEligibility(ctx context.Context, mpxn string, eligible bool) error
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
	eligible, cached, err := c.cache.GetEligibilityForMpxn(ctx, mpan)
	if err != nil {
		return false, err
	}

	if cached {
		return eligible, nil
	}

	eligible, err = c.gw.GetMeterpointEligibility(ctx, mpan, mprn, postcode)
	if err != nil {
		return false, err
	}

	err = c.cache.CacheEligibility(ctx, mpan, eligible)
	return eligible, err
}
