-include .env
export

SHELL := /bin/bash

BACKEND_DIR := backend
BUILD_DIR := builds
DOCKER_COMPOSE := docker compose
DEV_COMPOSE_FILE := ./infrastructure/docker/docker-compose.dev.yml
NO_DOCKER_COMPOSE_FILE := ./infrastructure/docker/docker-compose.dev.no_docker.yml
TESTS_COMPOSE_FILE := ./infrastructure/docker/docker-compose.tests.yml
OBS_COMPOSE_FILE := ./infrastructure/docker/docker-compose.observability.yml
TESTS_DIR := tests
COVERAGE_FILE := $(TESTS_DIR)/coverage.out
COVERAGE_HTML := $(TESTS_DIR)/coverage.html
FRONTEND_DIR := frontend
SONAR_SCANNER_IMAGE := sonarsource/sonar-scanner-cli
SONAR_HOST_URL ?= http://host.docker.internal:9000
CODE_REVIEW_GRAPH_BIN := code-review-graph
REPO_ROOT := $(CURDIR)
MIGRATE_DIRECTION ?= up

.PHONY: help tidy_and_verify cleanup run run_obs run_full stop logs migrate migrate_up migrate_down migrate_docker run_no_docker run_with_graph_watch run_no_docker_with_graph_watch run_dev_all run_init_sonar run_sonar_backend run_sonar_frontend run_sonar_all run_test_sonar test_integration_up test_integration_migrate test_integration test_integration_down run_graph_build run_graph_watch run_graph_status push_code run_test_coverage_backend run_test_coverage_once_backend run_test_coverage_frontend run_test_coverage_once_frontend run_test_goconvey lint lint_backend lint_frontend backup backup_list restore

help:
	@printf '%s\n' \
		'Available targets:' \
		'  make tidy_and_verify        - tidy and verify Go module dependencies in backend' \
		'  make cleanup                - clean Go build and module cache' \
		'  make run                    - run the main dev docker compose stack' \
		'  make run_obs                - run observability services (expects dev network or use run_full)' \
		'  make run_full               - run dev stack together with observability' \
		'  make stop                   - stop dev, observability, and sonar compose stacks' \
		'  make logs                   - tail logs from the main dev docker compose stack' \
		'  make migrate_up             - run database migrations up inside Docker' \
		'  make migrate_down           - run database migrations down inside Docker' \
		'  make migrate MIGRATE_DIRECTION=up - run migrate container with explicit direction' \
		'  make run_no_docker          - start PostgreSQL + Redis in Docker and run backend locally' \
		'  make run_with_graph_watch - run dev docker compose with code-review-graph watch' \
		'  make run_no_docker_with_graph_watch - run local backend with code-review-graph watch' \
		'  make run_dev_all            - run dev docker compose with all watchers' \
		'  make run_init_sonar         - alias for backend Sonar scan' \
		'  make run_sonar_backend      - run backend Sonar scan locally in docker' \
		'  make run_sonar_frontend     - run frontend Sonar scan locally in docker' \
		'  make run_sonar_all          - run backend and frontend Sonar scans' \
		'  make run_test_sonar         - run docker compose for SonarQube stack' \
		'  make test_integration_up    - start sonar_db + redis for integration tests (no sonarqube app)' \
		'  make test_integration_migrate - apply migrations against the learnflow_test database' \
		'  make test_integration       - run backend integration tests (-tags=integration) against real Postgres/Redis' \
		'  make test_integration_down  - stop the integration test stack' \
		'  make run_graph_build        - rebuild code-review-graph for the repo' \
		'  make run_graph_watch        - watch repo changes and auto-update graph' \
		'  make run_graph_status       - show current code-review-graph status' \
		'  make push_code MSG=...      - commit and push to current origin' \
		'  make run_test_coverage_backend - generate coverage report and watch Go files' \
		'  make run_test_coverage_frontend - generate coverage report and watch Vue files' \
		'  make run_test_goconvey      - run GoConvey for backend' \
		'  make lint                   - run backend and frontend linters' \
		'  make lint_backend           - run golangci-lint on backend' \
		'  make lint_frontend          - run npm lint on frontend' \
		'  make backup                 - run manual DB backup (pg_dump → MinIO)' \
		'  make backup_list            - list all backups stored in MinIO' \
		'  make restore FILE=...       - restore DB from backup file in MinIO'

tidy_and_verify:
	@echo 'Verifying and tidying backend module dependencies...'
	cd $(BACKEND_DIR) && go mod tidy && go mod verify

cleanup:
	go clean -cache
	go clean -modcache
	go clean -testcache

run:
	$(DOCKER_COMPOSE) -f '$(DEV_COMPOSE_FILE)' up --build

run_obs:
	$(DOCKER_COMPOSE) -f '$(OBS_COMPOSE_FILE)' up --build

run_full:
	$(DOCKER_COMPOSE) -f '$(DEV_COMPOSE_FILE)' -f '$(OBS_COMPOSE_FILE)' up --build

stop:
	$(DOCKER_COMPOSE) -f '$(DEV_COMPOSE_FILE)' down
	$(DOCKER_COMPOSE) -f '$(OBS_COMPOSE_FILE)' down
	$(DOCKER_COMPOSE) -f '$(TESTS_COMPOSE_FILE)' down

logs:
	$(DOCKER_COMPOSE) -f '$(DEV_COMPOSE_FILE)' logs -f

migrate:
	$(DOCKER_COMPOSE) -f '$(DEV_COMPOSE_FILE)' up --build 'migrate'

migrate_up:
	cd $(BACKEND_DIR) && go run ./cmd/migrate -steps=1 -direction=up

migrate_down:
	cd $(BACKEND_DIR) && go run ./cmd/migrate -steps=0 -direction=down

migrate_docker:
	$(DOCKER_COMPOSE) -f '$(DEV_COMPOSE_FILE)' up 'migrate'

run_no_docker:
	$(DOCKER_COMPOSE) -f '$(NO_DOCKER_COMPOSE_FILE)' up --build -d
	cd $(BACKEND_DIR) && go run ./cmd/api


run_graph_build:
	$(CODE_REVIEW_GRAPH_BIN) build --repo "$(REPO_ROOT)"

run_graph_watch:
	$(CODE_REVIEW_GRAPH_BIN) watch --repo "$(REPO_ROOT)"

run_graph_status:
	$(CODE_REVIEW_GRAPH_BIN) status --repo "$(REPO_ROOT)"

run_with_graph_watch:
	@set -e; \
	$(MAKE) run_graph_watch & \
	GRAPH_PID=$$!; \
	trap 'kill $$GRAPH_PID 2>/dev/null || true' EXIT INT TERM; \
	$(MAKE) run

run_no_docker_with_graph_watch:
	@set -e; \
	$(MAKE) run_graph_watch & \
	GRAPH_PID=$$!; \
	trap 'kill $$GRAPH_PID 2>/dev/null || true' EXIT INT TERM; \
	$(MAKE) run_no_docker

run_dev_all:
	@set -e; \
	$(MAKE) run_graph_watch & \
	GRAPH_PID=$$!; \
	$(MAKE) run_test_coverage_backend & \
	BACKEND_COVER_PID=$$!; \
	$(MAKE) run_test_coverage_frontend & \
	FRONTEND_COVER_PID=$$!; \
	trap 'kill $$GRAPH_PID $$BACKEND_COVER_PID $$FRONTEND_COVER_PID 2>/dev/null || true' EXIT INT TERM; \
	$(MAKE) run

run_init_sonar:
	@$(MAKE) run_sonar_backend

run_sonar_backend:
	@test -n "$(SONAR_TOKEN)" || (echo "No SONAR_TOKEN in environment or .env"; exit 1)
	@$(MAKE) run_test_coverage_once_backend
	@set -e; \
	envfile=$$(mktemp); \
	chmod 600 "$$envfile"; \
	trap 'rm -f "$$envfile"' EXIT INT TERM; \
	printf 'SONAR_HOST_URL=%s\nSONAR_TOKEN=%s\n' "$(SONAR_HOST_URL)" "$(SONAR_TOKEN)" > "$$envfile"; \
	docker run --rm \
		--env-file "$$envfile" \
		-v "$$(pwd):/usr/src" \
		-w /usr/src/backend \
		$(SONAR_SCANNER_IMAGE)

run_sonar_frontend:
	@test -n "$(SONAR_TOKEN)" || (echo "No SONAR_TOKEN in environment or .env"; exit 1)
	cd $(FRONTEND_DIR) && npm run test:unit:coverage
	@set -e; \
	envfile=$$(mktemp); \
	chmod 600 "$$envfile"; \
	trap 'rm -f "$$envfile"' EXIT INT TERM; \
	printf 'SONAR_HOST_URL=%s\nSONAR_TOKEN=%s\n' "$(SONAR_HOST_URL)" "$(SONAR_TOKEN)" > "$$envfile"; \
	docker run --rm \
		--env-file "$$envfile" \
		-v "$$(pwd):/usr/src" \
		-w /usr/src/frontend \
		$(SONAR_SCANNER_IMAGE)

run_sonar_all: run_sonar_backend run_sonar_frontend

run_test_sonar:
	$(DOCKER_COMPOSE) -f '$(TESTS_COMPOSE_FILE)' up --build

test_integration_up:
	$(DOCKER_COMPOSE) -f '$(TESTS_COMPOSE_FILE)' up -d --wait sonar_db redis

test_integration_migrate: test_integration_up
	$(DOCKER_COMPOSE) -f '$(TESTS_COMPOSE_FILE)' up --build 'migrate'

test_integration: test_integration_migrate
	cd $(BACKEND_DIR) && \
		env $$(grep -v '^#' ../.env.test | xargs) DB_HOST=localhost DB_PORT=5433 \
		go test -tags=integration ./...

test_integration_down:
	$(DOCKER_COMPOSE) -f '$(TESTS_COMPOSE_FILE)' down

push_code:
	@test -n "$$MSG" || (echo "No MSG. Use: make push_code MSG='description'"; exit 1)
	git add .
	git diff --cached --quiet && echo "Nothing to commit." || git commit -m "$$MSG"
	git push -u origin HEAD

# Example:
# make push_code MSG="fix: prometheus + graceful shutdown"

run_test_coverage_once_backend:
	mkdir -p $(TESTS_DIR)
	cd $(BACKEND_DIR) && \
		go clean -testcache && \
		go test ./... -coverpkg=./... -coverprofile="$(CURDIR)/$(COVERAGE_FILE)" && \
		go tool cover -func="$(CURDIR)/$(COVERAGE_FILE)" | tail -n 1 && \
		go tool cover -html="$(CURDIR)/$(COVERAGE_FILE)" -o "$(CURDIR)/$(COVERAGE_HTML)"

run_test_coverage_backend: run_test_coverage_once_backend
	fswatch -r --one-per-batch \
		--exclude '(\.git|vendor|node_modules|dist|tmp|coverage\.out|coverage\.html)$$' \
		--include '\.go$$' $(BACKEND_DIR) \
	| while read -r _; do \
		(cd $(BACKEND_DIR) && go test ./... -coverpkg=./... -coverprofile="$(CURDIR)/$(COVERAGE_FILE)" && go tool cover -html="$(CURDIR)/$(COVERAGE_FILE)" -o "$(CURDIR)/$(COVERAGE_HTML)"); \
	done

run_test_coverage_once_frontend:
	cd $(FRONTEND_DIR) && npm run test:unit:coverage

run_test_coverage_frontend: run_test_coverage_once_frontend
	fswatch -r --one-per-batch \
		--exclude '(\.git|node_modules|dist|coverage)$$' \
		--include '(\.vue|\.ts|\.tsx|\.js|\.jsx)$$' $(FRONTEND_DIR)/app $(FRONTEND_DIR)/e2e $(FRONTEND_DIR)/vitest.config.ts $(FRONTEND_DIR)/vite.config.ts \
	| while read -r _; do \
		(cd $(FRONTEND_DIR) && npm run test:unit:coverage); \
	done

lint: lint_backend lint_frontend

lint_backend: lint_repo_interfaces
	cd $(BACKEND_DIR) && golangci-lint run ./...

lint_verify:
	cd $(BACKEND_DIR) && golangci-lint config verify

# lint_repo_interfaces: golangci-lint has no linter for *missing* compile-time interface
# assertions (it catches bad patterns, not absent enforcement). Every repository.go must
# assert `var _ domain.XRepository = (*Repository)(nil)` so a method added to a domain
# interface without a matching Repository implementation fails at compile time, not at
# runtime via a type assertion panic. This grep-based check fills that gap until a real
# linter rule exists.
lint_repo_interfaces:
	@fail=0; \
	while IFS= read -r -d '' f; do \
		if ! grep -qE '_ [A-Za-z0-9_.]*Repository = \(\*Repository\)\(nil\)' "$$f"; then \
			echo "missing compile-time interface check: $$f"; \
			fail=1; \
		fi; \
	done < <(find $(BACKEND_DIR)/internal -path "*/repository/repository.go" -print0); \
	exit $$fail

lint_frontend:
	cd $(FRONTEND_DIR) && npm run lint

backup:
	docker exec learnflow_backup backup.sh

backup_list:
	docker exec learnflow_backup mc ls local/${BACKUP_BUCKET:-learnflow-backups}

restore:
	@test -n "$$FILE" || (echo "No FILE. Use: make restore FILE=learnflow_20240101_120000.sql.gz"; exit 1)
	@echo "$$FILE" | grep -qE '^learnflow_[0-9]{8}_[0-9]{6}\.sql\.gz$$' || (echo "Invalid FILE: expected format learnflow_YYYYMMDD_HHMMSS.sql.gz, got: $$FILE"; exit 1)
	docker exec learnflow_backup restore.sh "$$FILE"
