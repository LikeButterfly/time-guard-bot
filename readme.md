# Time Guard Bot

Telegram Bot для управления временем задач

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
