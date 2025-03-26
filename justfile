install_deps:
    pnpm install
    go mod download
    go mod verify

generate:
    go generate ./internal/pkg/routes/site
    node_modules/.bin/tailwindcss -i ./static/css/_styles.css -o ./static/css/styles.css

run_migrations:
    go run ./cmd/migrations

generate_db_types: run_migrations
    go tool github.com/sqlc-dev/sqlc/cmd/sqlc generate

ci_prepare:
    go generate ./internal/pkg/routes/site
    go tool github.com/sqlc-dev/sqlc/cmd/sqlc generate

cd_prepare: ci_prepare
    node_modules/.bin/tailwindcss -i ./static/css/_styles.css -o ./static/css/styles.css

build_docs: generate generate_db_types

_docs: build_docs
    go tool golang.org/x/pkgsite/cmd/pkgsite -http=:6060

docs:
    go tool github.com/air-verse/air -c .air.docs.toml

build: generate
    mkdir -p build
    go build -o build/portfolio ./cmd/main

cd_build: cd_prepare
    go build -tags=ci -o build/portfolio ./cmd/main

lint: build
    go tool -modfile=golangci-lint.mod github.com/golangci/golangci-lint/cmd/golangci-lint run

dev:
    go tool github.com/air-verse/air

clean:
    rm -rf build
    rm -rf node_modules

run: build
    ./build/portfolio

[confirm("Please make sure that DB_URL is set to a production database URL.\nPress ENTER to continue, use Ctrl+C to cancel > ")]
deploy: generate_db_types lint
    source .env
    fly deploy
