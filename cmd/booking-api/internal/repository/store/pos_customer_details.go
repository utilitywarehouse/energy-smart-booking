package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store/serialisers"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/go-operational/op"
)

const (
	prefixKeyCustomerDetails = "account_details"
)

var ErrPOSCustomerDetailsNotFound = errors.New("accounts details not found")

type AccountDetailsStore struct {
	r   *redis.Client
	ttl time.Duration
}

func NewAccountDetailsStore(r *redis.Client, ttl time.Duration) *AccountDetailsStore {
	return &AccountDetailsStore{r: r, ttl: ttl}
}

func (s *AccountDetailsStore) NewHealthCheck() func(*op.CheckResponse) {
	return func(cr *op.CheckResponse) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.r.Info(ctx).Err(); err != nil {
			cr.Unhealthy("cant connect to redis: "+err.Error(), "Troubleshoot redis",
				"account details cache cannot be cached")
			return
		}

		cr.Healthy("Redis is healthy")
	}
}

func (s *AccountDetailsStore) key(accountNumber string) string {
	return fmt.Sprintf("%s:%s", prefixKeyCustomerDetails, accountNumber)
}

func (s *AccountDetailsStore) SetAccountDetails(ctx context.Context, accountDetails models.PointOfSaleCustomerDetails) error {
	posDetails := serialisers.PointOfSaleCustomerDetails{}
	b, err := posDetails.Serialise(accountDetails)
	if err != nil {
		return err
	}

	return s.r.Set(ctx, s.key(accountDetails.AccountNumber), string(b), s.ttl).Err()
}

func (s *AccountDetailsStore) GetAccountDetails(ctx context.Context, accountNumber string) (*models.PointOfSaleCustomerDetails, error) {
	accountDetailsStr, err := s.r.Get(ctx, s.key(accountNumber)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrPOSCustomerDetailsNotFound
		}
		return nil, err
	}
	posDetails := serialisers.PointOfSaleCustomerDetails{}

	accountDetails, err := posDetails.Deserialise([]byte(accountDetailsStr))
	if err != nil {
		return nil, err
	}

	return &accountDetails, nil
}
