.DEFAULT_GOAL := build

.PHONY: all update frontend-update build types-build frontend-build backend-build publish backend-publish backend-publish-build db-apply db-diff db-inspect db-plan

all: update build

update: frontend-update

frontend-update:
	@echo "[frontend] Updating..."
	cd frontend && ncu --upgrade && npm update

build: types-build frontend-build backend-build

types-build:
	@echo "[types] Building..."
	cd backend/type_generation && GOEXPERIMENT=jsonv2 go run type_generation.go --typescript-output ../../frontend/src/generated.ts --postgres-output ../database/generated.sql

frontend-build:
	@echo "[frontend] Building..."
	cd frontend && npm run build

backend-build:
	@echo "[backend] Building..."
	#cd backend && go generate && GOOS=linux go build -a -ldflags="-s -w -buildid=" -installsuffix cgo -o ../service
	cd backend && go generate && GOEXPERIMENT=jsonv2 go build -o ../service

publish: backend-publish

backend-publish-build:
	@echo "[backend] Building for publish..."
	cd backend && podman build . --tag X

backend-publish: backend-publish-build
	@echo "[backend] Publishing..."
	podman tag X Y/X:latest && podman push Y/X:latest

db-apply:
	@echo "[database] Applying schema..."
	grep -i "CREATE EXTENSION" backend/database/generated.sql | psql "$(DB_URL)" -f - > /dev/null 2>&1 || true
	cd backend/database && atlas schema apply --env local --url "$(DB_URL)&search_path=public" --to file://generated.sql

db-diff:
	@echo "[database] Calculating diff..."
	cd backend/database && atlas schema diff --env local --from "$(DB_URL)&search_path=public" --to file://generated.sql

db-inspect:
	@echo "[database] Inspecting active schema..."
	cd backend/database && atlas schema inspect --env local --url "$(DB_URL)&search_path=public"

db-plan:
	@echo "[database] Planning schema changes..."
	cd backend/database && atlas schema apply --env local --url "$(DB_URL)&search_path=public" --to file://generated.sql --dry-run
