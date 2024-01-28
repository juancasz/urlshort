package redis_test

import (
	"context"
	"os"
	"strconv"
	"testing"
	"urlshort/internal/redis"
)

func TestSaveAndGet(t *testing.T) {
	run := os.Getenv("RUN_INTEGRATION_TESTS")
	if run != "true" {
		t.Skip("set RUN_INTEGRATION_TESTS to true to run this test")
	}

	expirationMinutes, err := strconv.Atoi(os.Getenv("REDIS_EXPIRATION_MINUTES"))
	if err != nil {
		t.Fatalf("REDIS_EXPIRATION_MINUTES not numeric: %s", err.Error())
	}

	storage := redis.New(&redis.Options{
		Host:              os.Getenv("REDIS_HOST"),
		Port:              os.Getenv("REDIS_PORT"),
		Username:          os.Getenv("REDIS_USERNAME"),
		Password:          os.Getenv("REDIS_PASSWORD"),
		ExpirationMinutes: expirationMinutes,
	})

	myKey := "my-key"
	myUrl := "http://www.google.com"
	if err = storage.Save(context.Background(), myKey, myUrl); err != nil {
		t.Fatalf("error was not expected but got: %s", err.Error())
	}

	myRetrievedUrl, err := storage.Get(context.Background(), myKey)
	if err != nil {
		t.Fatalf("error was not expected but got: %s", err.Error())
	}

	if myRetrievedUrl != myUrl {
		t.Fatalf("expected url %s but got %s", myUrl, myRetrievedUrl)
	}
}
