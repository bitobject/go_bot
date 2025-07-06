# API Архитектура

## Обзор

Система построена с использованием чистой архитектуры и лучших практик для production:

- **Gin** - HTTP фреймворк
- **JWT** - аутентификация
- **Rate Limiting** - защита от перебора
- **Structured Logging** - структурированное логирование
- **Graceful Shutdown** - корректное завершение
- **Health Checks** - проверка здоровья сервиса

## Структура проекта

```
internal/
├── api/
│   ├── handlers/          # HTTP обработчики
│   │   ├── admin.go       # Админ endpoint'ы
│   │   ├── health.go      # Health checks
│   │   └── webhook.go     # Telegram webhook
│   ├── middleware/        # Middleware
│   │   ├── auth.go        # JWT аутентификация
│   │   ├── logging.go     # Логирование
│   │   └── rate_limit.go  # Rate limiting
│   └── server.go          # Основной сервер
├── auth/
│   └── jwt.go            # JWT утилиты
├── bot/
│   └── handlers.go       # Telegram логика
├── config/
│   └── config.go         # Конфигурация
└── database/
    ├── models.go         # Модели БД
    ├── admin.go          # Админ сервис
    └── postgres.go       # Подключение к БД
```

## Endpoint'ы

### Health Checks
- `GET /health` - проверка здоровья сервиса
- `GET /ready` - проверка готовности к работе

### Admin API
- `POST /api/admin/login` - вход администратора
- `GET /api/admin/profile` - профиль администратора (требует JWT)
- `POST /api/admin/change-password` - смена пароля (требует JWT)

### Telegram
- `POST /api/webhook` - webhook от Telegram

## Аутентификация

### Вход администратора
```bash
curl -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"login": "admin", "password": "password"}'
```

Ответ:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "admin": {
    "id": 1,
    "login": "admin"
  }
}
```

### Использование JWT
```bash
curl -X GET http://localhost:8080/api/admin/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## Middleware

### Rate Limiting
- 200 запросов в минуту на IP
- Настраивается через `RATE_LIMIT_REQUESTS` и `RATE_LIMIT_WINDOW`

### Логирование
- Структурированные JSON логи
- Уровень логирования настраивается через `LOG_LEVEL`
- Логирует все HTTP запросы с метаданными

### Аутентификация
- Проверяет JWT токен в заголовке `Authorization`
- Добавляет `admin_id` и `admin_login` в контекст

## Конфигурация

Все настройки через environment variables:

```bash
# Скопируйте env.example в .env и настройте
cp env.example .env
```

### Обязательные переменные
- `DATABASE_URL` - строка подключения к PostgreSQL
- `TELEGRAM_TOKEN` - токен Telegram бота
- `JWT_SECRET` - секретный ключ для JWT (измените в production!)

### Опциональные переменные
- `APP_PORT` - порт сервера (по умолчанию 8080)
- `APP_HOST` - хост сервера (по умолчанию 0.0.0.0)
- `JWT_EXPIRATION` - время жизни JWT (по умолчанию 24h)
- `RATE_LIMIT_REQUESTS` - лимит запросов (по умолчанию 200)
- `RATE_LIMIT_WINDOW` - окно для rate limiting (по умолчанию 1m)
- `LOG_LEVEL` - уровень логирования (по умолчанию info)

## Безопасность

### JWT
- Использует HMAC-SHA256
- Содержит стандартные claims (iss, aud, iat, exp, nbf)
- Время жизни 24 часа
- Секретный ключ из environment

### Rate Limiting
- Защита от брутфорса
- Настраиваемые лимиты
- Отдельные лимиты для разных endpoint'ов

### Пароли
- Хеширование через bcrypt (cost 12)
- Защита от timing attacks
- Блокировка после 5 неудачных попыток

## Мониторинг

### Health Checks
- `/health` - общая проверка здоровья
- `/ready` - проверка готовности (включает проверку БД)

### Логирование
- JSON формат для парсинга
- Уровни: debug, info, warn, error
- Метаданные запросов (IP, User-Agent, время ответа)

## Запуск

```bash
# Установка зависимостей
go mod tidy

# Запуск
go run cmd/bot/main.go
```

## Production

### Рекомендации
1. Измените `JWT_SECRET` на уникальный секретный ключ
2. Настройте `DATABASE_URL` для production БД
3. Установите `LOG_LEVEL=info` или `LOG_LEVEL=warn`
4. Настройте reverse proxy (nginx) для SSL
5. Используйте Docker для развертывания

### Docker
```bash
# Сборка
docker build -t goooo-bot .

# Запуск
docker run -p 8080:8080 --env-file .env goooo-bot
``` 