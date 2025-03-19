build:
    go generate ./internal/pkg/routes/site
    mkdir -p build
    go build -o build/portfolio cmd/main/main.go

run: build
    ./build/portfolio
