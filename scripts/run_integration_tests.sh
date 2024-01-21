#!/bin/sh
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_EXPIRATION_MINUTES=60
export RUN_INTEGRATION_TESTS=true

docker compose -f docker/redis.yaml up -d
go test ./... -coverprofile cover.out && go tool cover -func cover.out
docker compose -f docker/redis.yaml down