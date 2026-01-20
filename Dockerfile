# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Build Go binary
COPY . .
RUN go build -o main main.go

# Download migrate tool
RUN apk add --no-cache curl tar
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.19.1/migrate.linux-amd64.tar.gz | tar xvz \
    && chmod +x migrate

# Run stage
FROM alpine:3.20
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate

# Copy config and scripts
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY db/migration ./migration

# Make executable
RUN chmod +x ./migrate ./start.sh ./wait-for.sh

EXPOSE 8080
ENTRYPOINT [ "/app/start.sh" ]
CMD ["./main"]