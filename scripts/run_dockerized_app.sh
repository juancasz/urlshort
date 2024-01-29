#!/bin/sh
docker compose -f docker/redis/app.yml up --build -d && docker compose -f docker/redis/app.yml logs -f
docker image prune --filter label=stage=builder