# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.26-alpine AS backend
RUN apk add --no-cache build-base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=1 go build -o /flash-cards ./cmd/server

# Stage 3: Minimal runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=backend /flash-cards /usr/local/bin/flash-cards
EXPOSE 8080
VOLUME ["/data"]
ENV DB_PATH=/data/flash-cards.db
ENTRYPOINT ["flash-cards"]
