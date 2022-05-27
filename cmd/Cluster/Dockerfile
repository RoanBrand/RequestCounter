# syntax=docker/dockerfile:1.3

# Build
FROM registry.semaphoreci.com/golang:1.18 as builder

WORKDIR /src

RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/cluster ./cmd/Cluster/*.go

# Deploy
FROM gcr.io/distroless/base-debian11

ARG PORT

COPY --from=builder ./bin/cluster ./cluster

EXPOSE ${PORT}

ENTRYPOINT ["./cluster"]
