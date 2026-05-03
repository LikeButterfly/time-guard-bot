# Time Guard Bot

Telegram Bot for managing task time tracking in team environments

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
  - [Running with Docker Compose](#running-with-docker-compose)
  - [Running from Source](#running-from-source)
  - [Running with Executable (Windows)](#running-with-executable-windows)
- [Available Commands](#available-commands)
- [API Documentation](#api-documentation)
- [Development](#development)

## Features

- ⏱️ Task time tracking with timers
- 🔒 Task locking mechanism
- 👥 Multi-user support in group chats
- 📊 Task status monitoring
- 🔌 REST API for external integrations
- 📚 Swagger documentation
- 🐳 Docker support

## Quick Start

### Prerequisites

1. **Telegram Bot Token**: Get it from [@BotFather](https://t.me/botfather)
2. **Redis**: Required for data storage

### Running with Docker Compose

The easiest way to run the bot:

1. Create `.env` file:
```bash
TELEGRAM_TOKEN=your_bot_token_here
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
API_ADDR=0.0.0.0:8080
```

2. Start the bot:
```bash
docker-compose up -d
```

3. Check logs:
```bash
docker-compose logs -f bot
```

4. Stop the bot:
```bash
docker-compose down
```

### Running from Source

1. Install Go 1.23 or later

2. Install Redis:
   - **Windows**: Download from [Redis releases](https://github.com/microsoftarchive/redis/releases) or use Docker
   - **Linux**: `sudo apt install redis-server`
   - **macOS**: `brew install redis`

3. Clone and setup:
```bash
git clone https://github.com/yourusername/time-guard-bot.git
cd time-guard-bot
go mod download
```

4. Create `.env` file:
```bash
TELEGRAM_TOKEN=your_bot_token_here
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
API_ADDR=0.0.0.0:8080
```

5. Start Redis (if not using Docker):
```bash
redis-server
```

6. Generate Swagger docs and run:
```bash
make swagger
make run
```

Or build and run separately:
```bash
make build
./time-guard-bot.exe  # Windows
./time-guard-bot      # Linux/macOS
```

### Running with Executable (Windows)

Perfect for quick deployment without Go installation.

#### Option 1: With Docker for Redis

1. Download the latest `time-guard-bot.exe` from [Releases](https://github.com/yourusername/time-guard-bot/releases)

2. Install Docker Desktop for Windows

3. Start Redis in Docker:
```powershell
docker run -d -p 6379:6379 --name redis redis:alpine
```

4. Create `.env` file in the same directory as the exe:
```
TELEGRAM_TOKEN=your_bot_token_here
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
API_ADDR=0.0.0.0:8080
```

5. Run the bot:
```powershell
.\time-guard-bot.exe
```

#### Option 2: With Native Redis

1. Download `time-guard-bot.exe` from [Releases](https://github.com/yourusername/time-guard-bot/releases)

2. Download Redis for Windows:
   - Get from [Redis releases](https://github.com/microsoftarchive/redis/releases)
   - Or use [Memurai](https://www.memurai.com/) (recommended for Windows)

3. Install and start Redis

4. Create `.env` file (same as Option 1)

5. Run:
```powershell
.\time-guard-bot.exe
```

#### Building Executable Yourself

If you want to build from source:

```bash
# Generate Swagger docs (optional, but recommended)
make swagger

# Build
make build

# The executable will be created as time-guard-bot.exe
```

Now you can copy `time-guard-bot.exe` and `.env` file to any Windows machine and run it

## Available commands

Task Management:

- `/add {task_name} [task_desc]` - Add a new task with a name and optional description
- `/delete {task_id}` - Delete a task by ID
- `/tasks` - List all tasks
- `/status [task_name]` - Show status of all tasks or a specific task
- `/lock {task_id} [reason]` - Lock a task, preventing it from being started
- `/unlock {task_id}` - Unlock a previously locked task

Time Tracking:

- `/{minutes} {task_name}` - Start a timer for a task (e.g., '/30 coding')
- `/cancel [task_name]` - Cancel specified timer (defaults to latest)

## API Documentation

The API is documented using Swagger. To access the Swagger UI:

1. Before launching the bot generate Swagger documentation:

```bash
make swagger
```

2. After launching the bot, open the Swagger UI in your browser:

```
http://localhost:8080/swagger/index.html
```

### API Endpoints

- `GET /api/task/status` - Get the status of a specific task
- `GET /api/task/list` - Get a list of all tasks

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TELEGRAM_TOKEN` | *required* | Bot token from @BotFather |
| `REDIS_ADDR` | `localhost:6379` | Redis server address |
| `REDIS_PASSWORD` | *(empty)* | Redis password (if required) |
| `REDIS_DB` | `0` | Redis database number |
| `API_ADDR` | `0.0.0.0:8080` | API server listen address |

**Important**: When running locally (not in Docker), you can use `:8080` for `API_ADDR`. When running in Docker, use `0.0.0.0:8080` to make the API accessible from outside the container

## Development

### Project Structure

```
time-guard-bot/
├── cmd/bot/            # Application entry point
├── internal/
│   ├── api/            # REST API server
│   ├── bot/            # Telegram bot logic
│   ├── helpers/        # Helper functions
│   ├── models/         # Data models
│   └── storage/        # Data storage layer
│       └── redis/      # Redis implementation
├── docs/swagger/       # Generated Swagger docs
├── Dockerfile          # Docker build configuration
├── docker-compose.yaml # Docker Compose setup
└── Makefile            # Build commands
```

## Project Building and Linter Configuration Guide

### Project preparation

1. **Deleting the cache of an old build**

```bash
go clean -cache
```

2. **Building a binary**

    The binary will be compiled into the root directory as time-guard-bot.exe:

```bash
make build
```

3. **Launching the bot**

```bash
make run
```

### Setting up and using golangci-lint

1. **Installing golangci-lint**

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

2. **Launching the linter**

    Runs all linters on all project files:

```bash
golangci-lint run
```

3. **Auto fix**

    Fixes automatically fixed issues:

```bash
golangci-lint run --fix
```

## Development recommendations

Before commit do:

- `golangci-lint run` - to check the code
- `make swagger` - to update the API documentation

### Before Committing

- ✅ Run `golangci-lint run` to check code quality
- ✅ Run `go test ./...` to ensure tests pass
- ✅ Run `make swagger` if you changed API endpoints
