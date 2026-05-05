# ─── Stage 1 : Frontend ───────────────────────────────────────────────────────
FROM node:20-alpine3.21 AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci --frozen-lockfile
COPY frontend/ .
RUN npm run build

# ─── Stage 2 : Backend ────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/api

# ─── Stage 3 : Image finale ───────────────────────────────────────────────────
FROM alpine:3.21
RUN apk --no-cache add ca-certificates

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /home/appuser/
COPY --from=backend-builder /app/main .
COPY --from=frontend-builder /app/frontend/dist ./static

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./main"]
