# Система администраторов

## Обзор

Система администраторов реализована с использованием лучших практик безопасности:

- **Хеширование паролей**: bcrypt с cost 12
- **Защита от брутфорса**: блокировка после 5 неудачных попыток
- **Case-insensitive логины**: использование CITEXT в PostgreSQL
- **Аудит**: отслеживание времени последнего входа и попыток входа

## Структура таблицы

```sql
CREATE TABLE admins (
    id SERIAL PRIMARY KEY,
    login CITEXT UNIQUE NOT NULL,
    hashed_password VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP WITH TIME ZONE,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## Использование

### Создание администратора

```go
adminService := database.NewAdminService(db)

// Создать нового администратора
admin, err := adminService.CreateAdmin("admin", "secure_password_123")
if err != nil {
    log.Printf("Failed to create admin: %v", err)
}
```

### Аутентификация

```go
// Аутентификация администратора
authenticatedAdmin, err := adminService.AuthenticateAdmin("admin", "password")
if err != nil {
    switch err {
    case database.ErrAdminNotFound:
        log.Println("Admin not found")
    case database.ErrInvalidCredentials:
        log.Println("Invalid credentials")
    case database.ErrAccountLocked:
        log.Println("Account is locked")
    case database.ErrAccountInactive:
        log.Println("Account is inactive")
    }
} else {
    log.Printf("Successfully authenticated: %s", authenticatedAdmin.Login)
}
```

### Управление аккаунтом

```go
// Обновить пароль
err = adminService.UpdateAdminPassword(adminID, "new_password")

// Деактивировать аккаунт
err = adminService.DeactivateAdmin(adminID)

// Активировать аккаунт
err = adminService.ActivateAdmin(adminID)

// Разблокировать аккаунт
err = adminService.UnlockAdmin(adminID)
```

### Генерация безопасного пароля

```go
// Сгенерировать случайный пароль
securePassword, err := database.GenerateSecurePassword(16)
if err != nil {
    log.Printf("Failed to generate password: %v", err)
} else {
    log.Printf("Generated password: %s", securePassword)
}
```

## Безопасность

### Хеширование паролей
- Используется bcrypt с cost 12 (рекомендуемое значение)
- Пароли никогда не хранятся в открытом виде
- Автоматическое обновление хеша при смене пароля

### Защита от брутфорса
- Максимум 5 неудачных попыток входа
- Блокировка на 15 минут после превышения лимита
- Автоматический сброс счетчика при успешном входе

### Дополнительные меры
- Case-insensitive логины (CITEXT)
- Отслеживание времени последнего входа
- Возможность деактивации аккаунтов
- Аудит неудачных попыток входа

## Миграция базы данных

При запуске приложения автоматически создаются все необходимые таблицы:

```go
// В database/postgres.go
err = db.AutoMigrate(&User{}, &Message{}, &Admin{})
```

## Примеры ошибок

```go
var (
    ErrAdminNotFound      = errors.New("admin not found")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrAccountLocked      = errors.New("account is locked")
    ErrAccountInactive    = errors.New("account is inactive")
)
```

## Рекомендации по использованию

1. **Пароли**: Используйте сложные пароли (минимум 12 символов, включая буквы, цифры и спецсимволы)
2. **Логины**: Используйте уникальные логины, не связанные с личной информацией
3. **Мониторинг**: Регулярно проверяйте логи неудачных попыток входа
4. **Ротация**: Периодически меняйте пароли администраторов
5. **Доступ**: Ограничьте доступ к административным функциям только необходимым пользователям 