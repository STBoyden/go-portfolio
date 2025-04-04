install_deps:
    pnpm install
    mkdir -p static/js
    cp node_modules/htmx.org/dist/htmx.min.js static/js
    cp node_modules/htmx-ext-preload/dist/preload.min.js static/js/htmx-preload.min.js
    cp node_modules/alpinejs/dist/cdn.min.js static/js/alpinejs.min.js
    go mod download

generate: install_deps
    go generate ./internal/pkg/routes/site
    node_modules/.bin/tailwindcss -i ./static/css/_styles.css -o ./static/css/styles.css

run_migrations:
    go run ./cmd/migrations

generate_db_types:
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

lint: generate
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
