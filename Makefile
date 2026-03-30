.PHONY: build test lint format check quality clean backend-check frontend-check

build:
	cd backend && go build ./...
	cd frontend && npm run build

test:
	cd backend && go test ./...
	cd frontend && npm run test

lint:
	cd backend && golangci-lint run ./...
	cd frontend && npx oxlint src/

format:
	cd backend && gofumpt -w .
	cd frontend && npx biome format --write src/

check: lint test build
	@echo "All checks passed."

quality:
	@echo "=== Quality Gate ==="
	@test -f LICENSE || { echo "ERROR: LICENSE missing. Fix: add MIT LICENSE file"; exit 1; }
	@! grep -rn "TODO\|FIXME\|HACK\|console\.log\|fmt\.Print" backend/ frontend/src/ 2>/dev/null | grep -v "node_modules" | grep -v "_test\.go" || { echo "ERROR: debug output or TODO found. Fix: remove before ship"; exit 1; }
	@! grep -rn "password=\|secret=\|api_key=\|sk-\|ghp_" backend/ frontend/src/ 2>/dev/null | grep -v '\$$' | grep -v "node_modules" || { echo "ERROR: hardcoded secrets. Fix: use env vars with no default"; exit 1; }
	@test ! -f PRD.md || ! grep -q "\[ \]" PRD.md || { echo "ERROR: unchecked acceptance criteria in PRD.md"; exit 1; }
	@echo "OK: automated quality checks passed"
	@echo "Manual checks required: README quickstart, demo GIF, input validation, ADR >=1"

backend-check:
	cd backend && gofumpt -w . && golangci-lint run ./... && go test ./... && go build ./...

frontend-check:
	cd frontend && npx oxlint src/ && npx biome format --write src/ && npm run test && npm run build

clean:
	cd backend && go clean ./...
	cd frontend && rm -rf dist node_modules
