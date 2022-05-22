# syntax=docker/dockerfile:1

# Build
FROM registry.semaphoreci.com/golang:1.18 as builder

WORKDIR /

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY cmd/RequestCounter/main.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o requestcounter ./main.go

# Deploy
FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=builder ./requestcounter ./requestcounter

EXPOSE 8023

USER requestcounter:requestcounter

ENTRYPOINT ["./requestcounter"]
