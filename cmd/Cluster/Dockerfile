# syntax=docker/dockerfile:1

# Build
FROM registry.semaphoreci.com/golang:1.18 as builder

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cluster ./cmd/Cluster/*.go

# Deploy
FROM gcr.io/distroless/base-debian11

ARG PORT
WORKDIR /

COPY --from=builder ./app/cluster ./cluster

EXPOSE ${PORT}

ENTRYPOINT ["./cluster"]
