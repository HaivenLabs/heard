FROM golang:1.26-alpine@sha256:28d89ee9cc0ff9fec75c82ca201e6bf7fdf9a679d4b7b24dfa04f2bb766bb468 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/worker ./cmd/worker

FROM alpine:3.22@sha256:14358309a308569c32bdc37e2e0e9694be33a9d99e68afb0f5ff33cc1f695dce AS api
RUN adduser -D -u 10001 heard
USER heard
COPY --from=build /out/api /usr/local/bin/heard-api
ENTRYPOINT ["/usr/local/bin/heard-api"]

FROM alpine:3.22@sha256:14358309a308569c32bdc37e2e0e9694be33a9d99e68afb0f5ff33cc1f695dce AS worker
RUN adduser -D -u 10001 heard
USER heard
COPY --from=build /out/worker /usr/local/bin/heard-worker
ENTRYPOINT ["/usr/local/bin/heard-worker"]
