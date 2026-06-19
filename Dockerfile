# ─── Stage 1 : Frontend ───────────────────────────────────────────────────────
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# ─── Stage 2 : Backend ────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS backend-builder
WORKDIR /app

ARG BUILD_VERSION=dev
ARG BUILD_COMMIT=unknown
ARG BUILD_DATE=unknown

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-s -w -X main.Version=${BUILD_VERSION} -X main.Commit=${BUILD_COMMIT} -X main.BuildDate=${BUILD_DATE}" \
  -o main ./cmd/api

# ─── Stage 3 : Image finale ───────────────────────────────────────────────────
# Alpine 3.22 — latest patched release, minimal CVE surface
FROM alpine:3.22

ARG BUILD_VERSION=dev
ARG BUILD_COMMIT=unknown
ARG BUILD_DATE=unknown

LABEL org.opencontainers.image.title="AniQuiz" \
      org.opencontainers.image.description="Anime music quiz game" \
      org.opencontainers.image.version="${BUILD_VERSION}" \
      org.opencontainers.image.revision="${BUILD_COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.source="https://github.com/LeYapson/aniquiz" \
      org.opencontainers.image.licenses="MIT"

# ca-certificates for TLS; tzdata for correct timezone; wget is BusyBox built-in (no extra package)
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

USER appuser
WORKDIR /home/appuser/

COPY --from=backend-builder --chown=appuser:appgroup /app/main .
COPY --from=frontend-builder --chown=appuser:appgroup /app/frontend/dist ./static

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./main"]
