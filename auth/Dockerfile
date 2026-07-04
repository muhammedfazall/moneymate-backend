
# STAGE 1: Base Builder (Shared Context)
FROM golang:1.26.4-alpine AS builder
WORKDIR /moneymate-backend

# STAGE 2: Local Development Environment
FROM builder AS dev

RUN go install github.com/air-verse/air@latest
CMD ["air", "-c", ".air.toml"]

# STAGE 3: Production Builder
FROM builder AS prod-builder
COPY auth/go.mod auth/go.sum ./auth/
WORKDIR /moneymate-backend/auth
RUN go mod download

WORKDIR /moneymate-backend
COPY shared/ ./shared/
COPY db/ ./db/
COPY auth/ ./auth/

WORKDIR /moneymate-backend/auth
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server ./cmd/main.go

# STAGE 4: Final Secure Production Image
FROM alpine:latest AS prod
WORKDIR /app/

# Install CA certificates (required if your app makes external HTTPS calls)
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user for security (Principle of Least Privilege)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy ONLY the compiled binary from Stage 3
COPY --from=prod-builder /server .

# Switch to the non-root user
USER appuser

# Document the port
EXPOSE 8081

# Run the raw binary
CMD ["./server"]