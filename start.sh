#!/bin/bash
export DB_HOST=localhost
export DB_PORT=5433
export DB_USER=user_db
export DB_PASS=password_db
export DB_NAME=rip_db
export MINIO_ENDPOINT=localhost:9000
export MINIO_ACCESS_KEY=minioadmin
export MINIO_SECRET_KEY=minioadmin123
export MINIO_BUCKET=images

go run cmd/main/main.go
