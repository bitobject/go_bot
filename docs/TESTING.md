# Тестирование API

## Обзор

Проект использует многоуровневый подход к тестированию с лучшими практиками:

- **Unit тесты** - тестирование отдельных функций
- **Integration тесты** - тестирование полного API с middleware
- **Benchmark тесты** - тестирование производительности
- **Coverage тесты** - измерение покрытия кода

## Структура тестов

```
internal/
├── api/
│   ├── handlers/
│   │   └── admin_test.go      # Unit тесты для admin handlers
│   └── integration_test.go    # Integration тесты для API
└── database/
    └── admin_test.go          # Unit тесты для admin service
```

## Запуск тестов

### Все тесты
```bash
make test
# или
go test -v ./...
```

### Unit тесты
```bash
make test-unit
# или
go test -v ./internal/api/handlers/ -run "^TestAdminHandler"
```

### Integration тесты
```bash
make test-integration
# или
go test -v ./internal/api/ -run "^TestAPI"
```

### Benchmark тесты
```bash
make test-benchmark
# или
go test -bench=. -benchmem ./...
```

### Тесты с покрытием
```bash
make test-coverage
# или
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Типы тестов

### 1. Unit тесты (`admin_test.go`)

Тестируют отдельные функции handlers без внешних зависимостей:

```go
func TestAdminHandler_Login_Success(t *testing.T) {
    // Arrange
    handler := setupTestHandler(t)
    admin := handler.createTestAdmin(t, "testadmin", "testpassword123")
    
    // Act
    loginReq := LoginRequest{Login: "testadmin", Password: "testpassword123"}
    // ... выполнение запроса
    
    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    assert.NotEmpty(t, response.Token)
}
```

**Особенности:**
- Используют in-memory SQLite базу данных
- Изолированные тесты без внешних зависимостей
- Быстрое выполнение
- Тестируют бизнес-логику

### 2. Integration тесты (`integration_test.go`)

Тестируют полный API с middleware и роутингом:

```go
func TestAPI_AdminLogin_Integration(t *testing.T) {
    // Arrange
    ts := setupTestServer(t)
    admin := ts.createTestAdmin(t, "testadmin", "testpassword123")
    
    // Act
    w := ts.makeRequest(t, "POST", "/api/admin/login", loginReq, nil)
    
    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    // ... проверки ответа
}
```

**Особенности:**
- Тестируют полный HTTP стек
- Включают middleware (auth, rate limiting, logging)
- Проверяют реальные HTTP запросы
- Тестируют интеграцию компонентов

### 3. Benchmark тесты

Измеряют производительность критических операций:

```go
func BenchmarkAdminHandler_Login_Success(b *testing.B) {
    handler := setupTestHandler(&testing.T{})
    // ... настройка
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // ... выполнение операции
    }
}
```

## Лучшие практики

### 1. Структура тестов (AAA Pattern)

```go
func TestFunction(t *testing.T) {
    // Arrange - подготовка данных
    handler := setupTestHandler(t)
    expected := "expected result"
    
    // Act - выполнение действия
    result := handler.SomeFunction()
    
    // Assert - проверка результата
    assert.Equal(t, expected, result)
}
```

### 2. Table Driven Tests

```go
func TestAdminHandler_Login_InvalidCredentials(t *testing.T) {
    testCases := []struct {
        name     string
        login    string
        password string
    }{
        {"wrong_password", "admin", "wrong"},
        {"wrong_login", "wrong", "password"},
        {"both_wrong", "wrong", "wrong"},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // тест для каждого случая
        })
    }
}
```

### 3. Тестовые данные

```go
// Создание тестовых данных
func (h *TestAdminHandler) createTestAdmin(t *testing.T, login, password string) *database.Admin {
    admin, err := h.adminService.CreateAdmin(login, password)
    require.NoError(t, err)
    return admin
}
```

### 4. Очистка после тестов

```go
func TestSomething(t *testing.T) {
    handler := setupTestHandler(t)
    defer handler.db.Migrator().DropTable(&database.Admin{}) // очистка
    
    // тест
}
```

## Тестовые сценарии

### Admin Login API

1. **Успешный вход**
   - Правильные логин/пароль
   - Возвращает JWT токен
   - Проверка валидности токена

2. **Неверные учетные данные**
   - Неверный пароль
   - Неверный логин
   - Оба неверные

3. **Некорректные запросы**
   - Отсутствующие поля
   - Неверный JSON
   - Пустой запрос

4. **Блокировка аккаунта**
   - Превышение лимита попыток
   - Неактивный аккаунт

### Admin Profile API

1. **Получение профиля**
   - С валидным JWT токеном
   - Без токена
   - С неверным токеном

### Admin Change Password API

1. **Смена пароля**
   - Успешная смена
   - Неверный текущий пароль
   - Слишком короткий новый пароль

### Health Checks

1. **Health endpoint**
   - Проверка статуса сервиса
   - Проверка метаданных

2. **Readiness endpoint**
   - Проверка готовности
   - Проверка подключения к БД

### Rate Limiting

1. **Превышение лимита**
   - Множественные запросы
   - Проверка блокировки
   - Проверка retry_after

## Настройка тестового окружения

### 1. In-Memory база данных

```go
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    
    err = db.AutoMigrate(&database.User{}, &database.Message{}, &database.Admin{})
    require.NoError(t, err)
    
    return db
}
```

### 2. Тестовая конфигурация

```go
func setupTestHandler(t *testing.T) *TestAdminHandler {
    config.LoadEnv()
    config.Init() // загружает настройки из .env
    
    db := setupTestDB(t)
    handler := NewAdminHandler(db)
    
    return &TestAdminHandler{...}
}
```

### 3. Мок объекты

```go
// Создание мок бота для тестов
bot := &tgbotapi.BotAPI{}
```

## Покрытие кода

### Запуск с покрытием
```bash
make test-coverage
```

### Анализ покрытия
```bash
# Открыть HTML отчет
open coverage.html

# Показать покрытие в консоли
go tool cover -func=coverage.out
```

### Целевое покрытие
- **Unit тесты**: 90%+ покрытие бизнес-логики
- **Integration тесты**: 100% покрытие API endpoint'ов
- **Error handling**: 100% покрытие обработки ошибок

## CI/CD интеграция

### GitHub Actions
```yaml
- name: Run tests
  run: make ci-test

- name: Build
  run: make ci-build
```

### Локальная проверка
```bash
# Полная проверка перед коммитом
make ci-test
make ci-build
```

## Отладка тестов

### Verbose режим
```bash
go test -v ./...
```

### Запуск конкретного теста
```bash
go test -v -run "TestAdminHandler_Login_Success"
```

### Параллельное выполнение
```bash
go test -parallel 4 ./...
```

### Race detection
```bash
go test -race ./...
```

## Рекомендации

1. **Пишите тесты для нового кода**
   - Каждая новая функция должна иметь тесты
   - Покрытие должно быть >90%

2. **Используйте правильные assertions**
   - `require` для критических проверок
   - `assert` для обычных проверок

3. **Изолируйте тесты**
   - Каждый тест должен быть независимым
   - Используйте `t.Parallel()` где возможно

4. **Тестируйте edge cases**
   - Ошибки сети
   - Некорректные данные
   - Граничные значения

5. **Поддерживайте тесты**
   - Обновляйте тесты при изменении кода
   - Рефакторите тесты вместе с кодом 