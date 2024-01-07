#!/bin/sh
export PORT=8080
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_EXPIRATION_MINUTES=60

docker compose -f docker/redis.yaml up --build -d

go run cmd/redis/main.go