# Base build image
FROM golang:1.12-alpine AS build_base

# Install some dependencies needed to build the project
RUN apk add bash ca-certificates git gcc g++ libc-dev
# Cache packages
RUN go get github.com/joho/godotenv/autoload \
    github.com/google/logger \
    github.com/google/uuid \
    github.com/Azure/azure-storage-blob-go/azblob \
    github.com/caarlos0/env

WORKDIR /go/src/app
COPY *.go ./
RUN go get
RUN go build -o azupload

#In this last stage, we start from a fresh Alpine image, to reduce the image size and not ship the Go compiler in our production artifacts.
FROM alpine AS azmarket
# We add the certificates to be able to verify remote weaviate instances
RUN apk add ca-certificates
WORKDIR /app
COPY --from=build_base /go/src/app/azupload /app/azupload
ENTRYPOINT ["/app/azupload"]
