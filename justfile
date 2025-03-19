install_tools_and_deps:
    pnpm install

generate: install_tools_and_deps
    go generate ./internal/pkg/routes/site
    node_modules/.bin/tailwindcss -i ./static/_styles.css -o ./static/styles.css

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
