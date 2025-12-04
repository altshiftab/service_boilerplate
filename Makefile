.DEFAULT_GOAL := all

.PHONY: all update frontend-update build sql-build typescript-build frontend-build backend-build publish backend-publish backend-publish-build db-apply db-diff db-inspect db-plan db-ensure run-test-server

all: update build

update: frontend-update

frontend-update:
	@echo "[frontend] Updating..."
	cd frontend && ncu --upgrade && npm update

build: frontend-build backend-build

sql-build:
	@echo "[sql] Building..."
	cd .sql_generation && GOEXPERIMENT=jsonv2 go run sql_generation.go --out generated.sql

typescript-build:
	@echo "[typescript] Building..."
	cd .typescript_generation && GOEXPERIMENT=jsonv2 go run typescript_generation.go --out ../frontend/src/scripts/generated.ts --domain "$$(jq -r '.domain' ../config.json)"

frontend-build: typescript-build
	@echo "[frontend] Building..."
	cd frontend && npm run build

backend-build: sql-build
	@echo "[backend] Building..."
	cd backend && go generate ./... && GOEXPERIMENT=jsonv2 go build -o ../service

publish: backend-publish

backend-publish-build:
	@echo "[backend] Building for publish..."
	cd backend && podman build --secret id=git_token,src=git_token.txt . --tag X

backend-publish: backend-publish-build
	@echo "[backend] Publishing..."
	podman tag X Y:latest && podman push Y:latest

db-apply: db-ensure
	@echo "[database] Applying schema..."
	cd .sql_generation && atlas schema apply --env local --url "$(DB_URL)&search_path=public" --to file://generated.sql

db-diff: db-ensure
	@echo "[database] Calculating diff..."
	cd .sql_generation && atlas schema diff --env local --from "$(DB_URL)&search_path=public" --to file://generated.sql

db-inspect: db-ensure
	@echo "[database] Inspecting active schema..."
	cd .sql_generation && atlas schema inspect --env local --url "$(DB_URL)&search_path=public"

db-plan: db-ensure
	@echo "[database] Planning schema changes..."
	cd .sql_generation && atlas schema apply --env local --url "$(DB_URL)&search_path=public" --to file://generated.sql --dry-run

db-ensure:
	@echo "[database] Ensuring database exists..."
	@DB_NAME=$$(printf "%s" "$(DB_URL)" | sed -E 's#.*://[^/]+/([^?]+).*#\1#'); \
	ADMIN_URL=$$(printf "%s" "$(DB_URL)" | sed -E 's#(.*://[^/]+)/[^?]+(.*)#\1/postgres\2#'); \
	psql "$$ADMIN_URL" -tAc "SELECT 1 FROM pg_database WHERE datname = '$$DB_NAME'" | grep -q 1 && echo "[database] Database '$$DB_NAME' exists." || (echo "[database] Creating database '$$DB_NAME'..." && psql "$$ADMIN_URL" -c "CREATE DATABASE \"$$DB_NAME\"")
	grep -i "CREATE EXTENSION" .sql_generation/generated.sql | psql "$(DB_URL)" -f - > /dev/null 2>&1 || true
