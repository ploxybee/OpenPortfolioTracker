BACKEND_DIR := backend
FRONTEND_DIR := frontend

.PHONY: dev api web test format

dev:
	@$(MAKE) -j2 api web

api:
	cd $(BACKEND_DIR) && go run ./cmd/api

web:
	cd $(FRONTEND_DIR) && npm run dev

test:
	cd $(BACKEND_DIR) && go test ./...
	cd $(FRONTEND_DIR) && npm run lint

format:
	cd $(BACKEND_DIR) && gofmt -w $$(find . -name '*.go')
	cd $(FRONTEND_DIR) && npm run format
