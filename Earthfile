VERSION 0.7
FROM golang:1.21-bookworm
WORKDIR /workspace

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
    proto/upload/v1alpha1/upload.proto
  SAVE ARTIFACT internal/gen AS LOCAL internal/gen
  SAVE ARTIFACT web/src/gen AS LOCAL web/src/gen

tidy:
  BUILD +tidy-go
  BUILD +lint-web

lint:
  BUILD +lint-go
  BUILD +lint-web

test:
  BUILD +test-go
  BUILD +test-web

build:
  BUILD +build-go
  BUILD +build-web

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

build-go:
  COPY . .
  COPY +build-web/dist ./web/dist
  RUN CGO_ENABLED=0 go build --ldflags '-s' -o bucketeer cmd/main.go
  SAVE ARTIFACT bucketeer AS LOCAL dist/bucketeer

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

tools:
  RUN curl -fsSL https://deb.nodesource.com/setup_21.x | bash -
  RUN apt install -y nodejs protobuf-compiler
  RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  RUN go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
  RUN npm install -g @bufbuild/protoc-gen-es @connectrpc/protoc-gen-connect-es