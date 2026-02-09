# Утилиты для работы с паролями

Этот пакет предоставляет функции для безопасной работы с паролями: хеширование, проверка и валидация.

## Функции

### PasswordHash

Создаёт bcrypt хеш пароля.

```go
hash, err := utils.PasswordHash("myPassword123")
if err != nil {
    // обработка ошибки
}
```

**Параметры:**
- `password` (string) - пароль в открытом виде

**Возвращает:**
- `string` - bcrypt хеш пароля
- `error` - ошибка, если не удалось создать хеш

**Примечание:**
- Использует стоимость (cost) из конфигурации (`config.Security.BcryptCost`)
- По умолчанию используется `bcrypt.DefaultCost` (обычно 10)
- Чем выше cost, тем безопаснее, но медленнее хеширование

---

### PasswordVerify

Проверяет, соответствует ли пароль хешу.

```go
isValid, err := utils.PasswordVerify("myPassword123", storedHash)
if err != nil {
    // обработка ошибки
}
if isValid {
    // пароль верный
}
```

**Параметры:**
- `password` (string) - пароль в открытом виде
- `hash` (string) - bcrypt хеш пароля (из базы данных)

**Возвращает:**
- `bool` - true, если пароль соответствует хешу
- `error` - ошибка, если произошла проблема при проверке

**Примечание:**
- Использует `bcrypt.CompareHashAndPassword`, который защищает от timing attacks

---

### PasswordValidate

Проверяет, соответствует ли пароль требованиям безопасности.

**Требования к паролю (как в PHP):**
- Минимум 10 символов
- Должен содержать строчные буквы (a-z)
- Должен содержать заглавные буквы (A-Z)
- Должен содержать цифры (0-9)
- Может содержать только буквы и цифры (без спецсимволов)

```go
err := utils.PasswordValidate("MyPassword123")
if err != nil {
    fmt.Println(err.Error())
    // Вывод: "password must be at least 10 characters"
    // или: "password must contain uppercase letters"
    // или: "password must contain lowercase letters"
    // или: "password must contain numbers"
    // или: "password must contain only letters and numbers"
}
```

**Параметры:**
- `password` (string) - пароль для проверки

**Возвращает:**
- `error` - nil, если пароль валиден, или описание ошибки

**Примеры ошибок:**
- `"password must be at least 10 characters"` - пароль слишком короткий
- `"password must contain uppercase letters"` - нет заглавных букв
- `"password must contain lowercase letters"` - нет строчных букв
- `"password must contain numbers"` - нет цифр
- `"password must contain only letters and numbers"` - есть недопустимые символы

---

### PasswordValidateWithConfirm

Проверяет пароль и его подтверждение.

```go
err := utils.PasswordValidateWithConfirm("MyPassword123", "MyPassword123")
if err != nil {
    fmt.Println(err.Error())
}
```

**Параметры:**
- `password` (string) - пароль
- `passwordConfirm` (string) - подтверждение пароля

**Возвращает:**
- `error` - nil, если пароль валиден и совпадает с подтверждением, или описание ошибки

**Примечание:**
- Сначала проверяет, что пароли совпадают
- Затем валидирует сам пароль

---

### PasswordValidateRegex

Проверяет пароль с помощью регулярного выражения (как в PHP).

```go
isValid, err := utils.PasswordValidateRegex("MyPassword123")
if err != nil {
    // обработка ошибки
}
if isValid {
    // пароль соответствует регулярному выражению
}
```

**Параметры:**
- `password` (string) - пароль для проверки

**Возвращает:**
- `bool` - true, если пароль соответствует регулярному выражению
- `error` - ошибка, если не удалось выполнить проверку

**Регулярное выражение:**
```
/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{10,}$/
```

**Примечание:**
- Это альтернативный способ проверки, используемый для совместимости с PHP
- Рекомендуется использовать `PasswordValidate`, так как он даёт более понятные сообщения об ошибках

---

## Примеры использования

### Полный цикл: создание и проверка пароля

```go
package main

import (
    "ctweb/internal/utils"
    "fmt"
)

func main() {
    // 1. Валидация пароля перед хешированием
    password := "MySecurePassword123"
    err := utils.PasswordValidate(password)
    if err != nil {
        fmt.Printf("Пароль не валиден: %v\n", err)
        return
    }

    // 2. Хеширование пароля
    hash, err := utils.PasswordHash(password)
    if err != nil {
        fmt.Printf("Ошибка при хешировании: %v\n", err)
        return
    }

    // 3. Сохранение хеша в базу данных
    // ... (код сохранения в БД)

    // 4. При входе: проверка пароля
    loginPassword := "MySecurePassword123"
    isValid, err := utils.PasswordVerify(loginPassword, hash)
    if err != nil {
        fmt.Printf("Ошибка при проверке: %v\n", err)
        return
    }

    if isValid {
        fmt.Println("Пароль верный, вход разрешён")
    } else {
        fmt.Println("Пароль неверный, вход запрещён")
    }
}
```

### Валидация при регистрации пользователя

```go
func RegisterUser(username, password, passwordConfirm string) error {
    // Проверяем пароль и его подтверждение
    err := utils.PasswordValidateWithConfirm(password, passwordConfirm)
    if err != nil {
        return fmt.Errorf("ошибка валидации пароля: %w", err)
    }

    // Хешируем пароль
    hash, err := utils.PasswordHash(password)
    if err != nil {
        return fmt.Errorf("ошибка при хешировании пароля: %w", err)
    }

    // Сохраняем пользователя в базу данных
    // ... (код сохранения)

    return nil
}
```

### Проверка пароля при входе

```go
func Login(username, password string) (*User, error) {
    // Получаем пользователя из базы данных
    user, err := userRepository.FindByLogin(username)
    if err != nil {
        return nil, fmt.Errorf("пользователь не найден: %w", err)
    }

    // Проверяем пароль
    isValid, err := utils.PasswordVerify(password, user.Password)
    if err != nil {
        return nil, fmt.Errorf("ошибка при проверке пароля: %w", err)
    }

    if !isValid {
        return nil, fmt.Errorf("неверный пароль")
    }

    return user, nil
}
```

---

## Сравнение с PHP

### PHP код:
```php
// Хеширование
$hash = password_hash($password, PASSWORD_DEFAULT);

// Проверка
$isValid = password_verify($password, $hash);

// Валидация
if(strlen($password) < 10) {
    return 'Password must be longer than or equal 10 characters';
}
if(!preg_match("/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{10,}$/", $password)) {
    return 'Password must be contains mixed case letters and numbers';
}
```

### Go код:
```go
// Хеширование
hash, err := utils.PasswordHash(password)

// Проверка
isValid, err := utils.PasswordVerify(password, hash)

// Валидация
err := utils.PasswordValidate(password)
```

---

## Безопасность

1. **bcrypt**: Используется алгоритм bcrypt, который является стандартом для хеширования паролей
2. **Соль (Salt)**: bcrypt автоматически добавляет уникальную соль к каждому паролю
3. **Cost**: Настраиваемая стоимость хеширования позволяет балансировать между безопасностью и производительностью
4. **Timing attacks**: `bcrypt.CompareHashAndPassword` защищает от timing attacks
5. **Никогда не храните пароли в открытом виде**: Всегда используйте хеширование перед сохранением в базу данных

---

## Конфигурация

Стоимость хеширования (cost) настраивается в `config/config.yaml`:

```yaml
security:
  bcrypt_cost: 10  # Рекомендуемое значение: 10-12
```

**Рекомендации по cost:**
- 10: хороший баланс скорости/безопасности (по умолчанию)
- 12: более безопасно, но медленнее
- 14+: очень безопасно, но может быть слишком медленно для высоконагруженных систем

---

## Зависимости

- `golang.org/x/crypto/bcrypt` - для хеширования и проверки паролей
- `ctweb/internal/config` - для получения конфигурации

