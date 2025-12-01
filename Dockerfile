FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .

RUN go mod tidy

RUN go build -o /grafana_git_sync ./cmd/grafana-git-sync

FROM scratch

WORKDIR /app

COPY --from=builder /grafana_git_sync /app/grafana_git_sync

# Health check endpoint on port 8080
EXPOSE 8080

# Note: Health check via HTTP endpoint :8080/healthz
# Docker/Kubernetes should check: http://localhost:8080/healthz

ENTRYPOINT ["/app/grafana_git_sync"]
