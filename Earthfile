VERSION 0.7
FROM golang:1.21-bookworm
WORKDIR /workspace

all:
  COPY (+build/bucketeer --GOARCH=amd64) ./dist/bucketeer-linux-amd64
  COPY (+build/bucketeer --GOARCH=arm64) ./dist/bucketeer-linux-arm64
  COPY (+build/bucketeer --GOOS=darwin --GOARCH=amd64) ./dist/bucketeer-darwin-amd64
  COPY (+build/bucketeer --GOOS=darwin --GOARCH=arm64) ./dist/bucketeer-darwin-arm64
  RUN cd dist && find . -type f -exec sha256sum {} \; >> ../checksums.txt
  SAVE ARTIFACT ./dist/bucketeer-linux-amd64 AS LOCAL dist/bucketeer-linux-amd64
  SAVE ARTIFACT ./dist/bucketeer-linux-arm64 AS LOCAL dist/bucketeer-linux-arm64
  SAVE ARTIFACT ./dist/bucketeer-darwin-amd64 AS LOCAL dist/bucketeer-darwin-amd64
  SAVE ARTIFACT ./dist/bucketeer-darwin-arm64 AS LOCAL dist/bucketeer-darwin-arm64
  SAVE ARTIFACT ./checksums.txt AS LOCAL dist/checksums.txt

build:
  ARG GOOS=linux
  ARG GOARCH=amd64
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  COPY +build-web/dist ./web/dist
  ARG CONSTANTS=github.com/bucket-sailor/bucketeer/internal/constants
  ARG TELEMETRY_URL=https://telemetry.bucket-sailor.com/api
  ARG VERSION=dev
  RUN --secret TELEMETRY_TOKEN=telemetry_token \
    CGO_ENABLED=0 go build --ldflags "-s \
      -X '${CONSTANTS}.TelemetryURL=${TELEMETRY_URL}' \
      -X '${CONSTANTS}.TelemetryToken=${TELEMETRY_TOKEN}' \
      -X '${CONSTANTS}.Version=${VERSION}'" \
    -o bucketeer cmd/main.go
  SAVE ARTIFACT bucketeer AS LOCAL dist/bucketeer-${GOOS}-${GOARCH}

tidy:
  BUILD +tidy-go
  BUILD +lint-web

lint:
  BUILD +lint-go
  BUILD +lint-web

test:
  BUILD +test-go
  BUILD +test-web

tidy-go:
  LOCALLY
  RUN go mod tidy
  RUN go fmt ./...

lint-go:
  FROM golangci/golangci-lint:v1.55.2
  WORKDIR /workspace
  COPY . .
  RUN mkdir -p web/dist \
    && echo 'hello' > web/dist/index.html
  RUN golangci-lint run --timeout 5m ./...

test-go:
  COPY go.* ./
  RUN go mod download
  COPY . .
  RUN mkdir -p web/dist \
    && echo 'hello' > web/dist/index.html
  RUN go test -coverprofile=coverage.out -v ./...
  SAVE ARTIFACT ./coverage.out AS LOCAL coverage.out

lint-web:
  FROM +deps-web
  COPY web .
  RUN npm run lint

test-web:
  FROM +deps-web
  COPY web .
  RUN npm run test

build-web:
  FROM +deps-web
  COPY web .
  RUN npm run build
  SAVE ARTIFACT dist AS LOCAL web/dist

deps-web:
  FROM +tools
  COPY web/package.json web/package-lock.json ./
  RUN npm install

generate:
  FROM +tools
  COPY proto/ ./proto
  RUN mkdir -p internal/gen web/src/gen
  RUN protoc -I=proto/ \
    --go_out=internal/gen \
    --go_opt=paths=source_relative \
    --connect-go_out=internal/gen \
    --connect-go_opt=paths=source_relative \
    --es_out=web/src/gen \
    --es_opt=target=ts \
    --connect-es_out=web/src/gen \
    --connect-es_opt=target=ts,import_extension=none \
    proto/filesystem/v1alpha1/filesystem.proto \
    proto/telemetry/v1alpha1/telemetry.proto \
    proto/upload/v1alpha1/upload.proto
  SAVE ARTIFACT internal/gen AS LOCAL internal/gen
  SAVE ARTIFACT web/src/gen AS LOCAL web/src/gen

tools:
  RUN curl -fsSL https://deb.nodesource.com/setup_21.x | bash -
  RUN apt install -y nodejs protobuf-compiler
  RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  RUN go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
  RUN npm install -g @bufbuild/protoc-gen-es @connectrpc/protoc-gen-connect-es