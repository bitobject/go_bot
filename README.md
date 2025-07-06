# Goooo Telegram Bot

Telegram bot built with Go, PostgreSQL, and Docker Compose.

## Features

- Telegram bot that responds with greetings
- PostgreSQL database for data persistence
- Nginx reverse proxy with security headers
- Docker Compose for easy deployment
- Health checks for all services

## Prerequisites

- Docker and Docker Compose installed
- Telegram Bot Token (get from [@BotFather](https://t.me/BotFather))

## Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd goooo
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env and add your TELEGRAM_TOKEN
   ```

3. **Start all services**
   ```bash
   make up
   # or
   docker-compose up -d
   ```

4. **Check status**
   ```bash
   make status
   # or
   docker-compose ps
   ```

## Project Structure

```
goooo/
├── cmd/bot/main.go          # Application entry point
├── internal/
│   ├── bot/handlers.go      # Telegram bot handlers
│   ├── config/config.go     # Configuration management
│   └── database/postgres.go # Database connection
├── nginx/
│   ├── nginx.conf          # Main nginx configuration
│   └── conf.d/default.conf # Server configuration
├── docker-compose.yml      # Docker Compose services
├── Dockerfile             # Go application container
├── init.sql              # Database initialization
├── Makefile              # Development commands
└── README.md             # This file
```

## Services

### App (Go Telegram Bot)
- **Port**: 8080 (internal)
- **Health Check**: `http://localhost:8080/health`
- **Environment**: TELEGRAM_TOKEN, DATABASE_URL

### PostgreSQL
- **Port**: 5432
- **Database**: goooo
- **User**: goooo_user
- **Password**: goooo_password
- **Health Check**: pg_isready

### Nginx
- **Ports**: 80, 443
- **Features**: Reverse proxy, security headers, rate limiting
- **Health Check**: HTTP on port 80

## Make Commands

```bash
make help      # Show all available commands
make up        # Start all services
make down      # Stop all services
make logs      # Show logs from all services
make status    # Show service status
make clean     # Remove all containers and volumes
make build     # Build Docker images
make restart   # Restart all services

# Development
make dev       # Start in development mode
make dev-build # Build and start in development mode

# Database
make db-shell  # Connect to PostgreSQL shell
make db-backup # Create database backup

# Application
make app-shell # Connect to app container

# Nginx
make nginx-reload # Reload nginx configuration
```

## Environment Variables

Create a `.env` file with the following variables:

```env
# Required
TELEGRAM_TOKEN=your_telegram_token_here

# Optional (used by docker-compose)
POSTGRES_DB=goooo
POSTGRES_USER=goooo_user
POSTGRES_PASSWORD=goooo_password
```

## Development

### Local Development
```bash
# Install dependencies
go mod tidy

# Run locally (requires PostgreSQL)
go run ./cmd/bot/main.go
```

### Docker Development
```bash
# Start services
make dev

# View logs
make logs-app

# Connect to database
make db-shell
```

## Production Deployment

1. **Set production environment variables**
2. **Use production Docker images**
3. **Configure SSL certificates for Nginx**
4. **Set up monitoring and logging**
5. **Configure backups for PostgreSQL**

## Security Features

- Nginx security headers
- Rate limiting (10 requests/second)
- Hidden file access denied
- Gzip compression
- Health checks for all services

## Troubleshooting

### Check service logs
```bash
make logs-app      # App logs
make logs-postgres # Database logs
make logs-nginx    # Nginx logs
```

### Restart services
```bash
make restart
```

### Clean and rebuild
```bash
make clean
make build
make up
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License 