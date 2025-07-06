# Goooo Bot

Telegram бот с админ-панелью, построенный на Go с использованием лучших практик для production.

## Архитектура

- **Gin** - HTTP фреймворк
- **GORM** - ORM для PostgreSQL
- **JWT** - аутентификация администраторов
- **Rate Limiting** - защита от перебора
- **Structured Logging** - структурированное логирование
- **Graceful Shutdown** - корректное завершение

## Быстрый старт

### 1. Настройка окружения

```bash
# Скопируйте пример конфигурации
cp env.example .env

# Отредактируйте .env файл
nano .env
```

### 2. Запуск базы данных

```bash
# Используя Docker Compose
docker-compose up -d postgres

# Или установите PostgreSQL локально
```

### 3. Создание первого администратора

```bash
# Создайте первого администратора
go run scripts/create_admin.go -login=admin -password=secure_password_123
```

### 4. Запуск приложения

```bash
# Установка зависимостей
go mod tidy

# Запуск
go run cmd/bot/main.go
```

### 5. Тестирование

```bash
# Запустите тесты API
./test_admin_login.sh
```

## API Endpoints

### Health Checks
- `GET /health` - проверка здоровья сервиса
- `GET /ready` - проверка готовности к работе

### Admin API
- `POST /api/admin/login` - вход администратора
- `GET /api/admin/profile` - профиль администратора (требует JWT)
- `POST /api/admin/change-password` - смена пароля (требует JWT)

### Telegram
- `POST /api/webhook` - webhook от Telegram

## Примеры использования

### Вход администратора

```bash
curl -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"login": "admin", "password": "secure_password_123"}'
```

### Получение профиля (с JWT токеном)

```bash
curl -X GET http://localhost:8080/api/admin/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Смена пароля

```bash
curl -X POST http://localhost:8080/api/admin/change-password \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"current_password": "old_password", "new_password": "new_secure_password"}'
```

## Конфигурация

### Обязательные переменные окружения

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
- 200 запросов в минуту на IP
- Защита от брутфорса
- Настраиваемые лимиты

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

## Разработка

### Структура проекта

```
internal/
├── api/              # HTTP API
│   ├── handlers/     # Обработчики запросов
│   ├── middleware/   # Middleware
│   └── server.go     # Основной сервер
├── auth/             # Аутентификация
├── bot/              # Telegram логика
├── config/           # Конфигурация
└── database/         # Работа с БД
```

### Добавление новых endpoint'ов

1. Создайте handler в `internal/api/handlers/`
2. Добавьте роут в `internal/api/server.go`
3. Добавьте middleware если нужно

## Лицензия

MIT 