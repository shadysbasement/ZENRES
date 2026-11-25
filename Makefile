# Zevra developer tasks. Requires Node 22+, Go 1.25+, Python 3.12+ (Docker for the full stack).
.PHONY: install dev dev-public dev-app dev-auth test test-web test-go test-worker lint lint-palette build up down seed-pilot enroll-admin

install:
	npm install
	cd services && go mod download
	cd services/intel-worker && python3 -m pip install -e ".[dev]"

dev-public: ; npm run dev:public
dev-app:    ; npm run dev:app
# Local auth service without Postgres (ephemeral):
dev-auth:
	cd services && STORE=memory COOKIE_SECURE=false ADMIN_GATE_SECRET=devgate INTERNAL_API_SECRET=devinternal go run ./auth-api/cmd/server

test: test-web test-go test-worker lint-palette
test-web:    ; npm test
test-go:     ; cd services && go vet ./... && go test ./...
test-worker: ; cd services/intel-worker && python3 -m pytest -q

lint-palette: ; node scripts/check-palette.mjs
lint: lint-palette ; npm run check

build: ; npm run build

up:   ; docker compose up --build -d
down: ; docker compose down
seed-pilot: ; docker compose run --rm seed

# make enroll-admin HANDLE=zv-founder-01 KEY=founder.asc
enroll-admin:
	cd services && go run ./auth-api/cmd/admin-enroll -handle $(HANDLE) -key $(abspath $(KEY))
