# syntax=docker/dockerfile:1

# Build
FROM registry.semaphoreci.com/golang:1.18 as builder

WORKDIR /

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY cmd/Cluster/main.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cluster ./main.go

# Deploy
FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=builder ./cluster ./cluster

EXPOSE 8084

ENTRYPOINT ["./cluster"]
