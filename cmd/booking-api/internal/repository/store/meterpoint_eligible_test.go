package store_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/cache"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
)

func SetupRedisTestContainer(ctx context.Context) (testcontainers.Container, error) {
	// Ryuk is used to clean up environments which can interfere with some CI environments where it is done automatically
	// See: https://golang.testcontainers.org/features/garbage_collector/#ryuk
	if err := os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true"); err != nil {
		return nil, err
	}

	req := testcontainers.ContainerRequest{
		Image:        "redis:alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	return redisC, err
}

func GetRedisTestContainerAddr(ctx context.Context, container testcontainers.Container) (string, error) {
	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		return "", err
	}

	return endpoint, err
}

func Test_MeterpointEligibleStore_CacheEligibility(t *testing.T) {
	ctx := context.Background()

	container, err := SetupRedisTestContainer(ctx)
	if err != nil {
		t.Fatalf("could not set up redis test container: %s", err.Error())
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	containerAddr, err := GetRedisTestContainerAddr(ctx, container)
	if err != nil {
		t.Fatalf("could not get redis test container address: %s", err.Error())
	}
	meterpointEligibleStore := store.NewMeterpointEligible(redis.NewClient(&redis.Options{Addr: containerAddr}), 6*time.Hour)

	type inputParams struct {
		mpan      string
		eligible  bool
		expiresAt time.Time
	}

	type testSetup struct {
		description string
		input       inputParams
		output      error
	}

	testCases := []testSetup{
		{
			description: "should cache a meterpoint eligible record",
			input: inputParams{
				mpan:     "mpan-1",
				eligible: true,
			},
			output: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := meterpointEligibleStore.SetEligibilityForMpxn(ctx, tc.input.mpan, "", tc.input.eligible)
			if err != tc.output {
				t.Fatalf("error output does not match, expected: %s | actual: %s", tc.output, err)
			}
		})
	}
}
func Test_MeterpointEligibleStore_GetEligibilityForMpxn(t *testing.T) {
	ctx := context.Background()

	container, err := SetupRedisTestContainer(ctx)
	if err != nil {
		t.Fatalf("could not set up redis test container: %s", err.Error())
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	containerAddr, err := GetRedisTestContainerAddr(ctx, container)
	if err != nil {
		t.Fatalf("could not get redis test container address: %s", err.Error())
	}
	meterpointEligibleStore := store.NewMeterpointEligible(redis.NewClient(&redis.Options{Addr: containerAddr}), 6*time.Hour)

	type inputParams struct {
		mpan          string
		eligible      bool
		skipInsertion bool
	}

	type testOutput struct {
		resbool bool
		reserr  error
	}

	type testSetup struct {
		description string
		input       inputParams
		output      testOutput
	}

	testCases := []testSetup{
		{
			description: "should cache and retrieve a meterpoint eligible record",
			input: inputParams{
				mpan:     "mpxn-1",
				eligible: true,
			},
			output: testOutput{
				resbool: true,
				reserr:  nil,
			},
		},
		{
			description: "should not cache and get a NotFound error when attempting retrieval",
			input: inputParams{
				mpan:          "mpxn-2",
				eligible:      true,
				skipInsertion: true,
			},
			output: testOutput{
				resbool: false,
				reserr:  cache.ErrNotFound,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if !tc.input.skipInsertion {
				err := meterpointEligibleStore.SetEligibilityForMpxn(ctx, tc.input.mpan, "", tc.input.eligible)
				if err != nil {
					t.Fatal("error when caching eligibility")
				}
			}
			res, err := meterpointEligibleStore.GetEligibilityForMpxn(ctx, tc.input.mpan, "")
			if err != tc.output.reserr {
				t.Fatalf("error output does not match, expected: %s | actual: %s", tc.output.reserr, err)
			}
			if res != tc.output.resbool {
				t.Fatalf("eligibility output does not match, expected: %t | actual: %t", tc.output.resbool, res)
			}
		})
	}
}
