FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/worker ./cmd/worker

FROM alpine:3.22 AS api
RUN adduser -D -u 10001 heard
USER heard
COPY --from=build /out/api /usr/local/bin/heard-api
ENTRYPOINT ["/usr/local/bin/heard-api"]

FROM alpine:3.22 AS worker
RUN adduser -D -u 10001 heard
USER heard
COPY --from=build /out/worker /usr/local/bin/heard-worker
ENTRYPOINT ["/usr/local/bin/heard-worker"]
