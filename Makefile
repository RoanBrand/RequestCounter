# Force to use buildkit for all images and for docker-compose to invoke
# Docker via CLI (otherwise buildkit isn't used for those images)
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

build:
	docker compose build

run:
	docker compose up

rebuild-run:
	docker compose up --build

test:
	go test -race ./cmd/Cluster/...

bench:
	go test -race -bench=. ./internal/db/...
