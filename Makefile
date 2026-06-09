.PHONY: install-tools migrate-up migrate-down migrate-test-up migrate-test-down generate-types generate-mocks

DATABASE_URL     ?= postgres://postgres:postgres@localhost:5432/research_events?sslmode=disable
TEST_DATABASE_URL ?= postgres://postgres:postgres@localhost:5433/research_events_test?sslmode=disable

GOOSE := $(shell go env GOPATH)/bin/goose

# Install goose once with: make install-tools
install-tools:
	go install github.com/pressly/goose/v3/cmd/goose@v3.27.1

migrate-up:
	$(GOOSE) -dir backend/migrations postgres "$(DATABASE_URL)" up

migrate-down:
	$(GOOSE) -dir backend/migrations postgres "$(DATABASE_URL)" down

migrate-test-up:
	$(GOOSE) -dir backend/migrations postgres "$(TEST_DATABASE_URL)" up

migrate-test-down:
	$(GOOSE) -dir backend/migrations postgres "$(TEST_DATABASE_URL)" down

generate-types:
	cd frontend && pnpm openapi-typescript ../specs/openapi.yaml -o src/types/api.ts

generate-mocks:
	cd backend && PATH="$$(go env GOPATH)/bin:$$PATH" go generate ./...
