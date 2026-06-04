.PHONY: migrate-up migrate-down generate-types generate-mocks

migrate-up:
	cd backend && goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	cd backend && goose -dir migrations postgres "$(DATABASE_URL)" down

generate-types:
	cd frontend && pnpm openapi-typescript ../specs/openapi.yaml -o src/types/api.ts

generate-mocks:
	cd backend && go generate ./...
