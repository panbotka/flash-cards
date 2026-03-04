#!/bin/bash
set -e
cd frontend && npm run build && cd ..
CGO_ENABLED=1 go build -o ./bin/flash-cards ./cmd/server
APP_PORT=3011 ./bin/flash-cards
