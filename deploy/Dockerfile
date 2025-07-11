# Этап 1: Сборка
# Используем официальный образ Go как основу для сборки.
# Указываем конкретную версию для воспроизводимости сборок.
FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые зависимости для сборки, включая CGO.
# CA-сертификаты нужны для того, чтобы Go-модули могли скачиваться по HTTPS.
RUN apk add --no-cache gcc musl-dev ca-certificates

# Устанавливаем рабочую директорию внутри контейнера.
WORKDIR /app

# Копируем файлы go.mod и go.sum для загрузки зависимостей.
# Этот слой кэшируется, если зависимости не менялись.
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной исходный код приложения.
COPY . .

# Собираем приложение.
# -ldflags "-w -s" - убирает отладочную информацию, уменьшая размер бинарника.
# CGO_ENABLED=0 - отключает CGO для создания статически скомпонованного бинарника.
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /go-bot ./cmd/bot

# Этап 2: Финальный образ
# Используем минимальный базовый образ scratch, который не содержит ничего лишнего.
FROM scratch

# Копируем CA-сертификаты из образа сборщика.
# Это необходимо для выполнения HTTPS-запросов из приложения.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем скомпилированный бинарник из образа сборщика.
COPY --from=builder /go-bot /go-bot

# Указываем порт, который будет слушать приложение.
EXPOSE 8080

# Команда для запуска приложения при старте контейнера.
ENTRYPOINT ["/go-bot"]
