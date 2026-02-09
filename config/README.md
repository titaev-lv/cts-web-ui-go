# Configuration

Конфигурационные файлы приложения.

## Файлы

- **config.yaml** - Основной конфигурационный файл (не коммитится в git, создаётся из примера)
- **config.example.yaml** - Пример конфигурации (коммитится в git)

## Быстрый старт

1. Скопируйте пример конфигурации:
   ```bash
   cp config/config.example.yaml config/config.yaml
   ```

2. Отредактируйте `config/config.yaml` и укажите свои значения:
   - Параметры подключения к БД
   - Секретные ключи (JWT, Session, CSRF)
   - Настройки сервера

3. Запустите приложение - конфигурация загрузится автоматически.

## Переменные окружения

Конфигурация также может быть переопределена через переменные окружения с префиксом `CT_`:

```bash
# Примеры
export CT_SERVER_PORT=8080
export CT_DATABASE_MYSQL_HOST=mysql.example.com
export CT_SECURITY_JWT_SECRET=my-secret-key
```

## Структура конфигурации

См. `config.example.yaml` для полной структуры. Основные секции:

- **server** - Настройки HTTP сервера
- **database** - Настройки подключения к БД
- **security** - Секретные ключи и настройки безопасности
- **logging** - Настройки логирования
- **daemon** - Настройки демона
- **app** - Общие настройки приложения

## Использование в коде

```go
import "ctweb/internal/config"

// Загрузка конфигурации (обычно в main.go)
cfg, err := config.Load("")
if err != nil {
    log.Fatal(err)
}

// Использование
port := config.Get().Server.Port
dsn := config.Get().GetMySQLDSN()
```

