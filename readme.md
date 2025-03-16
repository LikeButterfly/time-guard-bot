# TimeGuardBot

- Telegram Bot для управления временем задач

----------------------------

## Руководство по сборке проекта и настройке pre-commit

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

### Настройка pre-commit

1. **Установка pre-commit**

```bash
pip install pre-commit
```

2. **Установка хуков**

```bash
pre-commit install
```

3. **Запуск хуков**

Запускает все хуки на всех файлах:

```bash
pre-commit run -a
```
