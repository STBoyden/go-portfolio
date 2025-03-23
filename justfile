install_deps:
    pnpm install
    go mod download
    go mod verify

generate:
    go generate ./internal/pkg/routes/site
    node_modules/.bin/tailwindcss -i ./static/css/_styles.css -o ./static/css/styles.css

run_migrations:
    go run ./cmd/migrations/main.go

generate_db_types: run_migrations
    go tool github.com/sqlc-dev/sqlc/cmd/sqlc generate

build: generate
    mkdir -p build
    go build -o build/portfolio cmd/main/main.go

dev:
    go tool github.com/air-verse/air

clean:
    rm -rf build
    rm -rf node_modules

run: build
    ./build/portfolio
