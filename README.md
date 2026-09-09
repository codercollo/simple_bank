# Simple Bank

**A production-oriented REST API for managing bank accounts and transactions, built with Go.**

Simple Bank is a backend banking service designed to demonstrate clean API architecture, database-driven account management, authentication, session handling, automated testing, and containerized deployment.

## Features

* User authentication and login
* Bank account management
* Account and transaction APIs
* Secure session/token handling
* PostgreSQL persistence
* Type-safe database access with SQLC
* RESTful HTTP API
* Database migrations
* Docker and Docker Compose support
* Automated CI/CD
* AWS deployment support

## Stack

**Go · PostgreSQL 15 · SQLC · Docker · GitHub Actions · AWS ECR**

## Project Structure

```text
simple_bank/
├── api/                # HTTP handlers, routes, and API tests
├── db/                 # Database queries, schema, and migrations
├── token/              # Authentication and session tokens
├── util/               # Configuration and shared utilities
├── .github/workflows/  # CI/CD workflows
├── Dockerfile
├── docker-compose.yaml
├── Makefile
├── main.go
└── sqlc.yaml
```

## Quick Start

### Prerequisites

* Go 1.25+
* PostgreSQL 15+
* Docker & Docker Compose

Clone the repository:

```bash
git clone https://github.com/codercollo/simple_bank.git
cd simple_bank
```

Start the database and application:

```bash
docker compose up --build
```

Or run the application locally:

```bash
go mod download
make migrate
go run main.go
```

## Configuration

Configure the application through environment variables.

Typical configuration includes:

```env
DB_SOURCE=postgresql://user:password@localhost:5432/simple_bank
SERVER_ADDRESS=0.0.0.0:8080
TOKEN_SYMMETRIC_KEY=your-secret-key
```

See the environment files in the repository for deployment-specific configuration.

## Development

Run tests:

```bash
go test ./...
```

Build the application:

```bash
go build -o simple_bank .
```

Generate database code with SQLC:

```bash
sqlc generate
```

## Deployment

The project includes Docker and GitHub Actions workflows for automated builds and AWS deployment.

The deployment pipeline can:

1. Build the application containers
2. Build the database image
3. Load deployment secrets
4. Push images to **Amazon ECR**
5. Deploy the application to AWS

## API

The API provides endpoints for authentication and banking operations such as account management and transactions.

Example login request:

```http
POST /login
Content-Type: application/json

{
  "username": "john",
  "password": "password"
}
```

Authentication is handled through signed tokens and server-side session management.

## Goals

Simple Bank focuses on practical backend engineering principles:

* Clean separation of API, database, and authentication layers
* Strong typing with Go and SQLC
* Secure authentication and session management
* Reliable database migrations
* Automated testing
* Reproducible containerized environments
* CI/CD-ready infrastructure

## License

MIT
