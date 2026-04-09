# Frontend
FROM node:22-alpine AS frontend
WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm ci 2>/dev/null || npm install
COPY web/ ./
RUN npm run build

# Go binary (embeds dist/)
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY main.go ./
COPY --from=frontend /dist ./dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /aviation-brief .

# Runtime — small, non-root, fits K8s / GHCR
FROM alpine:latest
RUN adduser -D -u 65532 -g '' appuser
WORKDIR /home/appuser
COPY --from=builder /aviation-brief ./aviation-brief
USER appuser
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["./aviation-brief"]
