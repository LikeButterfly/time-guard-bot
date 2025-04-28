.PHONY: build run swagger
.SILENT:

build:
	go build -o time-guard-bot.exe ./cmd/bot

run: build
	./time-guard-bot.exe

swagger:
	go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g internal/api/docs.go -output ./docs/swagger --parseDependency
