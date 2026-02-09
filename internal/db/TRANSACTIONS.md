# Работа с транзакциями в базе данных

**Дата:** 14 декабря 2024

## Что такое транзакция?

**Транзакция** - это группа SQL запросов, которые выполняются как единое целое. Либо все запросы выполняются успешно, либо все откатываются (отменяются).

### Зачем нужны транзакции?

**Пример проблемы без транзакции:**
```
1. Создаём пользователя → УСПЕХ
2. Добавляем пользователя в группу → ОШИБКА (группа не существует)
Результат: пользователь создан, но без группы - данные неконсистентны!
```

**С транзакцией:**
```
1. Начинаем транзакцию
2. Создаём пользователя → УСПЕХ
3. Добавляем пользователя в группу → ОШИБКА
4. Откатываем транзакцию
Результат: пользователь НЕ создан, данные консистентны!
```

## Основные функции

### 1. BeginTransaction() - Начало транзакции

```go
tx, err := db.BeginTransaction()
if err != nil {
    return err
}
```

**Что делает:**
- Получает отдельное соединение из пула
- Начинает новую транзакцию на этом соединении
- Все последующие запросы будут выполняться в этой транзакции

### 2. CommitTransaction() - Подтверждение транзакции

```go
err = db.CommitTransaction(tx)
if err != nil {
    return err
}
```

**Что делает:**
- Сохраняет все изменения в базу данных
- Завершает транзакцию
- Освобождает соединение обратно в пул

### 3. RollbackTransaction() - Откат транзакции

```go
err = db.RollbackTransaction(tx)
if err != nil {
    return err
}
```

**Что делает:**
- Отменяет все изменения, сделанные в транзакции
- Завершает транзакцию
- Освобождает соединение

### 4. GetLastInsertID() - Получение ID вставленной записи

```go
result, err := tx.Exec("INSERT INTO users ...")
userID, err := db.GetLastInsertID(result)
```

**Что делает:**
- Возвращает ID последней вставленной записи
- Используется после INSERT запросов

### 5. GetRowsAffected() - Количество изменённых строк

```go
result, err := tx.Exec("UPDATE users SET ...")
rows, err := db.GetRowsAffected(result)
```

**Что делает:**
- Возвращает количество изменённых/удалённых строк
- Используется для проверки, что запрос действительно что-то изменил

## Паттерн использования (рекомендуемый)

```go
func CreateUserWithGroups(name, email string, groupIDs []int) error {
    // 1. Начинаем транзакцию
    tx, err := db.BeginTransaction()
    if err != nil {
        return err
    }

    // 2. Обязательно откатываем при ошибке
    defer func() {
        if err != nil {
            db.RollbackTransaction(tx)
        }
    }()

    // 3. Выполняем запросы
    result, err := tx.Exec("INSERT INTO USER ...", name, email)
    if err != nil {
        return err // defer выполнит Rollback
    }

    userID, err := db.GetLastInsertID(result)
    if err != nil {
        return err
    }

    // 4. Ещё запросы
    for _, groupID := range groupIDs {
        _, err = tx.Exec("INSERT INTO USERS_GROUP ...", userID, groupID)
        if err != nil {
            return err // defer выполнит Rollback
        }
    }

    // 5. Если всё ОК, подтверждаем
    err = db.CommitTransaction(tx)
    if err != nil {
        return err
    }

    // 6. Важно: обнуляем err, чтобы defer не откатил транзакцию
    err = nil
    return nil
}
```

## Важные правила

### ✅ ДЕЛАТЬ:

1. **Всегда используйте defer для Rollback**
   ```go
   defer func() {
       if err != nil {
           db.RollbackTransaction(tx)
       }
   }()
   ```

2. **Обнуляйте err после успешного Commit**
   ```go
   err = db.CommitTransaction(tx)
   if err != nil {
       return err
   }
   err = nil // Чтобы defer не откатил
   ```

3. **Проверяйте ошибки после каждого запроса**
   ```go
   _, err = tx.Exec("...")
   if err != nil {
       return err // Автоматически откатится через defer
   }
   ```

### ❌ НЕ ДЕЛАТЬ:

1. **Не забывайте вызывать Commit или Rollback**
   - Если не вызвать, соединение останется заблокированным

2. **Не используйте tx после Commit/Rollback**
   - Транзакция закрыта, tx больше не валиден

3. **Не вкладывайте транзакции напрямую**
   - В Go нет вложенных транзакций
   - Используйте функции, которые принимают tx как параметр

## Примеры из реального кода

### Создание пользователя с группами

```go
func CreateUser(login, password, email string, groupIDs []int) error {
    tx, err := db.BeginTransaction()
    if err != nil {
        return err
    }
    defer func() {
        if err != nil {
            db.RollbackTransaction(tx)
        }
    }()

    // Создаём пользователя
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
    result, err := tx.Exec(
        "INSERT INTO USER (LOGIN, PASSWORD, EMAIL, ACTIVE) VALUES (?, ?, ?, ?)",
        login, hashedPassword, email, 1,
    )
    if err != nil {
        return err
    }

    userID, err := db.GetLastInsertID(result)
    if err != nil {
        return err
    }

    // Добавляем в группы
    for _, groupID := range groupIDs {
        _, err = tx.Exec(
            "INSERT INTO USERS_GROUP (UID, GID) VALUES (?, ?)",
            userID, groupID,
        )
        if err != nil {
            return err
        }
    }

    err = db.CommitTransaction(tx)
    if err != nil {
        return err
    }

    err = nil
    return nil
}
```

### Обновление с проверкой

```go
func UpdateUser(userID int, newName string) error {
    tx, err := db.BeginTransaction()
    if err != nil {
        return err
    }
    defer func() {
        if err != nil {
            db.RollbackTransaction(tx)
        }
    }()

    // Обновляем
    result, err := tx.Exec(
        "UPDATE USER SET NAME = ? WHERE ID = ?",
        newName, userID,
    )
    if err != nil {
        return err
    }

    // Проверяем, что пользователь существовал
    rows, err := db.GetRowsAffected(result)
    if err != nil {
        return err
    }

    if rows == 0 {
        err = errors.New("user not found")
        return err
    }

    err = db.CommitTransaction(tx)
    if err != nil {
        return err
    }

    err = nil
    return nil
}
```

## Сравнение с PHP кодом

| PHP | Go |
|-----|-----|
| `$DB->startTransaction()` | `db.BeginTransaction()` |
| `$DB->commitTransaction()` | `db.CommitTransaction(tx)` |
| `$DB->rollbackTransaction()` | `db.RollbackTransaction(tx)` |
| `$DB->getLastID()` | `db.GetLastInsertID(result)` |

## Дополнительные ресурсы

- [Go database/sql Transactions](https://go.dev/doc/database/transactions)
- [MySQL Transactions](https://dev.mysql.com/doc/refman/8.0/en/commit.html)

---

*Транзакции - важный инструмент для обеспечения целостности данных!*

