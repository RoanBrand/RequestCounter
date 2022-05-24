# syntax=docker/dockerfile:1

# Build
FROM registry.semaphoreci.com/golang:1.18 as builder

WORKDIR /app

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o requestcounter ./cmd/RequestCounter/main.go

# Deploy
FROM gcr.io/distroless/base-debian11

ARG PORT
WORKDIR /

COPY --from=builder ./app/requestcounter ./requestcounter

EXPOSE ${PORT}

ENTRYPOINT ["./requestcounter"]
