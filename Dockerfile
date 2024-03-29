FROM golang:latest AS base_builder
LABEL maintainer="Charles Tenorio <charles.tenorio.dev@gmail.com>"

WORKDIR /myapp/

COPY ["go.mod", "go.sum", "./"]

RUN go mod download


### Build Go
FROM base_builder AS builder

WORKDIR /myapp/

COPY . .

ARG PROJECT_VERSION=1 CI_COMMIT_SHORT_SHA=1
RUN go build -ldflags="-s -w -X 'main.VERSION=$PROJECT_VERSION' -X main.COMMIT=$CI_COMMIT_SHORT_SHA" -o app cmd/worker/main.go


### Build Docker Image
FROM alpine:3.17

WORKDIR /app/

COPY --from=builder ["/myapp/app", "./"]

ENTRYPOINT ["./app"]

#export PROJECT_VERSION=$(cat $(pwd)/VERSION)
#export CI_COMMIT_SHORT_SHA=$(git rev-parse --short HEAD) ou pegar a $CI_COMMIT_SHORT_SHA do gitlab
#docker build --build-arg PROJECT_VERSION=$(cat $(pwd)/VERSION) --build-arg CI_COMMIT_SHORT_SHA=$(git rev-parse --short HEAD) -t acragentesvirtuaisdev.azurecr.io/plataforma-worker-financial-message-counter:$(cat $(pwd)/VERSION) -t acragentesvirtuaisdev.azurecr.io/plataforma-worker-financial-message-counter:latest . && docker compose up -d
