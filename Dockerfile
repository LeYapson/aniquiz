# --- Étape 1 : Build du Frontend ---
FROM node:20-alpine as frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# --- Étape 2 : Build du Backend Go ---
FROM golang:1.26-alpine as backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main ./cmd/api

# --- Étape 3 : Image Finale ---
FROM alpine:latest
WORKDIR /root/
# On récupère le binaire Go
COPY --from=backend-builder /app/main .
# On récupère les fichiers statiques du front pour que Go puisse les servir si besoin
COPY --from=frontend-builder /app/frontend/dist ./static
EXPOSE 8080
CMD ["./main"]