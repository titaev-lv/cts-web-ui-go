# Internal Package

Внутренние пакеты приложения, не предназначенные для внешнего использования.

## Структура

- **config/** - Загрузка и управление конфигурацией приложения
- **controllers/** - HTTP handlers (Gin controllers)
- **db/** - Подключение к базе данных и управление соединениями
- **dto/** - Data Transfer Objects (Request/Response модели)
- **logger/** - Система логирования
- **middleware/** - HTTP middleware (auth, security, logging)
- **models/** - Доменные модели данных
- **repositories/** - Слой доступа к данным (database operations)
- **services/** - Бизнес-логика приложения
- **utils/** - Вспомогательные утилиты (password hashing, validation, sanitization)

