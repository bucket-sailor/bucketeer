VERSION 0.7
FROM golang:1.21-bookworm
WORKDIR /workspace

generate:
  FROM +tools
  COPY proto/ ./proto
  RUN mkdir -p gen web/src/gen
  RUN protoc -I=proto/ \
    --go_out=gen \
    --go_opt=paths=source_relative \
    --connect-go_out=gen \
    --connect-go_opt=paths=source_relative \
    --es_out=web/src/gen \
    --es_opt=target=ts \
    --connect-es_out=web/src/gen \
    --connect-es_opt=target=ts,import_extension=none \
    proto/filesystem/v1alpha1/filesystem.proto \
    proto/upload/v1alpha1/upload.proto
  SAVE ARTIFACT gen AS LOCAL gen
  SAVE ARTIFACT web/src/gen AS LOCAL web/src/gen

tidy:
  LOCALLY
  RUN go mod tidy
  RUN go fmt ./...

lint:
  FROM golangci/golangci-lint:v1.55.2
  WORKDIR /workspace
  COPY . ./
  RUN golangci-lint run --timeout 5m ./...

test:
  COPY go.* ./
  RUN go mod download
  COPY . .
  RUN go test -coverprofile=coverage.out -v ./...
  SAVE ARTIFACT ./coverage.out AS LOCAL coverage.out

tools:
  RUN curl -fsSL https://deb.nodesource.com/setup_21.x | bash -
  RUN apt install -y nodejs protobuf-compiler
  RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  RUN go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
  RUN npm install -g @bufbuild/protoc-gen-es @connectrpc/protoc-gen-connect-es