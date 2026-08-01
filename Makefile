BACKEND_DIR := backend
FRONTEND_DIR := frontend

.PHONY: dev api web db-migrate sqlc test format

dev:
	@$(MAKE) -j2 api web

api:
	cd $(BACKEND_DIR) && go run ./cmd/api

web:
	cd $(FRONTEND_DIR) && npm run dev

db-migrate:
	cd $(BACKEND_DIR) && go run ./cmd/migrate

sqlc:
	cd $(BACKEND_DIR)/internal/model && sqlc generate

test:
	cd $(BACKEND_DIR) && go test ./...
	cd $(FRONTEND_DIR) && npm run lint

format:
	cd $(BACKEND_DIR) && gofmt -w $$(find . -name '*.go')
	cd $(FRONTEND_DIR) && npm run format
