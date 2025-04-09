# TimeGuardBot

- Telegram Bot для управления временем задач

----------------------------

## Руководство по сборке проекта и настройке линтера

### Подготовка проекта

1. **Удаление кэша старой сборки**

```bash
go clean -cache
```

2. **Сборка бинарника**

    Бинарник будет собран в директорию .bin/bot:

```bash
make build
```

3. **Запуск бота**

```bash
make run
```

### Настройка и использование golangci-lint

1. **Установка golangci-lint**

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

2. **Запуск линтера**

    Запускает все линтеры на всех файлах проекта:

```bash
golangci-lint run
```

3. **Автоматическое исправление проблем**

    Исправляет автоматически исправляемые проблемы:

```bash
golangci-lint run --fix
```
