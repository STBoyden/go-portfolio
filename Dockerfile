ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
RUN apt-get update && apt-get install nodejs npm -y
RUN npm install -g pnpm
COPY package.json pnpm-lock.yaml ./
RUN pnpm install
COPY go.mod go.sum justfile ./
RUN node_modules/.bin/just install_deps
COPY . .
RUN node_modules/.bin/just cd_build

FROM debian:bookworm

RUN apt-get update && apt-get install ca-certificates -y
RUN update-ca-certificates
COPY --from=builder /usr/src/app/build/portfolio /usr/local/bin/
CMD ["portfolio"]
