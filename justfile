set dotenv-load := true
set shell := ["bash", "-cu"]

default:
    @just --list

# Start PostgreSQL, the Go API, and Vite through devenv.
dev:
    devenv up

# Build and start PostgreSQL, the Go API, and nginx through Docker Compose.
docker:
    POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-postgres}" docker compose -f docker/docker-compose.yml up --build
