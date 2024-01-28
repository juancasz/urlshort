#!/bin/sh
export HOST=http://localhost:8080
export PORT=8080
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_EXPIRATION_MINUTES=60

docker compose -f docker/redis/redis.yaml up -d

go run cmd/redis/main.go