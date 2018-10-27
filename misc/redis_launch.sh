#!/usr/bin/env bash
docker run --name qabric-redis -p 6379:6379 -d redis redis-server --appendonly yes