# syntax=docker/dockerfile:1-experimental

# Build
FROM registry.semaphoreci.com/golang:1.18 as builder

WORKDIR /app

COPY . ./

RUN --mount=type=cache,target=~/.cache/go-build \
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o requestcounter ./cmd/RequestCounter/*.go

# Deploy
FROM gcr.io/distroless/base-debian11

ARG PORT
WORKDIR /

COPY --from=builder ./app/requestcounter ./requestcounter

EXPOSE ${PORT}

ENTRYPOINT ["./requestcounter"]
