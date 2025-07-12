# --- Этап 1: Сборщик (Builder) ---
# Используем конкретную версию Go на базе Alpine для консистентности и малого размера.
FROM golang:1.24-alpine AS builder

# Устанавливаем рабочую директорию.
WORKDIR /src

# 1. КОПИРУЕМ ТОЛЬКО ФАЙЛЫ ЗАВИСИМОСТЕЙ
# Этот слой кэшируется и пересобирается только при изменении go.mod или go.sum.
COPY go.mod go.sum ./

# 2. СКАЧИВАЕМ ЗАВИСИМОСТИ
# Этот шаг занимает много времени, но теперь он будет кэшироваться.
RUN go mod download

# Копируем весь остальной исходный код
COPY . .

# Собираем основной бинарник приложения
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/app ./cmd/bot

# Собираем бинарник для healthcheck
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/healthcheck ./cmd/healthcheck

# --- Финальный этап ---
# Используем пустой образ scratch для максимальной безопасности и минимального размера
FROM scratch

# Копируем CA-сертификаты из builder'а. Они необходимы для выполнения
# HTTPS-запросов из нашего приложения (например, к API Telegram).
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем скомпилированное приложение и healthcheck-клиент
COPY --from=builder /bin/app /app
COPY --from=builder /bin/healthcheck /healthcheck

# Создаем непривилегированного пользователя с UID 10001.
# USER 10001 не создает пользователя, а указывает Docker запустить
# контейнер от имени пользователя с этим UID. Это работает в scratch.
USER 10001

# Указываем, какой порт слушает приложение (только для информации, не открывает порт).
EXPOSE 8080

# Встроенная проверка состояния. Использует наш скомпилированный клиент.
# Это надежнее, чем внешние утилиты, особенно в scratch-образе.
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
  CMD ["/healthcheck", "http://localhost:8080/health"]

# Команда для запуска нашего приложения.
ENTRYPOINT ["/app"]