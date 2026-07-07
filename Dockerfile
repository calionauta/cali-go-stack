# syntax=docker/dockerfile:1.7
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    go tool templ generate && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /usr/local/bin/app ./cmd/web/

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /usr/local/bin/app /app/app
EXPOSE 8080
ENTRYPOINT ["/app/app"]
