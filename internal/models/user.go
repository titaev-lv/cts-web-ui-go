// Package models содержит структуры данных (модели) для работы с базой данных.
// Этот файл содержит модель User с поддержкой групп пользователей.
package models

import (
	"time"
)

// User представляет пользователя системы.
//
// Структура соответствует таблице USER в базе данных MySQL.
// Пользователь может принадлежать к нескольким группам через таблицу USERS_GROUP.
//
// Поля:
//   - ID: уникальный идентификатор пользователя
//   - Login: логин для входа (уникальный)
//   - Password: хеш пароля (bcrypt, не отправляется в JSON)
//   - Email: email адрес пользователя
//   - Active: активен ли пользователь (true = активен, false = заблокирован)
//   - Name: имя пользователя
//   - LastName: фамилия пользователя
//   - Token: токен для "Remember Me" (хранится в БД, используется в cookie)
//   - Timezone: временная зона пользователя (например, "UTC", "Europe/Moscow")
//   - DateCreate: дата создания записи
//   - DateModify: дата последнего изменения (может быть nil)
//   - TimestampX: временная метка последнего входа/активности (может быть nil)
//   - UserCreated: ID пользователя, создавшего эту запись (может быть nil)
//   - UserModify: ID пользователя, изменившего эту запись (может быть nil)
//   - Groups: список групп, к которым принадлежит пользователь (загружается отдельно)
//
// Пример использования:
//
//	user := models.User{
//	    Login:    "admin",
//	    Email:    "admin@example.com",
//	    Active:   true,
//	    Name:     "Admin",
//	    LastName: "User",
//	}
type User struct {
	ID         int        `json:"id" db:"ID"`                    // Уникальный идентификатор
	Login      string     `json:"login" db:"LOGIN"`              // Логин (уникальный)
	Password   string     `json:"-" db:"PASSWORD"`                // Хеш пароля (не отправляется в JSON)
	Email      string     `json:"email" db:"EMAIL"`               // Email адрес
	Active     bool       `json:"active" db:"ACTIVE"`             // Активен ли пользователь (true = активен)
	Name       string     `json:"name" db:"NAME"`                 // Имя
	LastName   string     `json:"last_name" db:"LAST_NAME"`       // Фамилия
	Token      string     `json:"-" db:"TOKEN"`                  // Токен для "Remember Me" (не отправляется в JSON)
	Timezone   string     `json:"timezone" db:"TIMEZONE"`         // Временная зона (например, "UTC")
	DateCreate time.Time  `json:"date_create" db:"DATE_CREATE"`  // Дата создания
	DateModify *time.Time `json:"date_modify,omitempty" db:"DATE_MODIFY"` // Дата изменения (может быть nil)
	TimestampX *time.Time `json:"timestamp_x,omitempty" db:"TIMESTAMP_X"` // Временная метка последней активности (может быть nil)
	UserCreated *int      `json:"user_created,omitempty" db:"USER_CREATED"` // ID создателя (может быть nil)
	UserModify  *int      `json:"user_modify,omitempty" db:"USER_MODIFY"`   // ID изменившего (может быть nil)

	// Groups - список ID групп, к которым принадлежит пользователь
	// Загружается отдельно через JOIN с таблицей USERS_GROUP
	// В JSON будет представлен как массив чисел: [1, 2, 3]
	Groups []int `json:"groups,omitempty" db:"-"` // Группы (не из БД напрямую, загружается через JOIN)
}

// GetFullName возвращает полное имя пользователя (Имя + Фамилия).
//
// Возвращает:
//   - string: полное имя, например "Иван Иванов"
//
// Пример использования:
//
//	fullName := user.GetFullName()
//	// Результат: "Иван Иванов"
func (u *User) GetFullName() string {
	if u.LastName != "" {
		return u.LastName + " " + u.Name
	}
	return u.Name
}

// IsActive проверяет, активен ли пользователь.
//
// Возвращает:
//   - bool: true если пользователь активен, false если заблокирован
//
// Пример использования:
//
//	if !user.IsActive() {
//	    return errors.UnauthorizedError("User is blocked")
//	}
func (u *User) IsActive() bool {
	return u.Active
}

// HasGroup проверяет, принадлежит ли пользователь к указанной группе.
//
// Параметры:
//   - groupID: ID группы для проверки
//
// Возвращает:
//   - bool: true если пользователь принадлежит к группе, false иначе
//
// Пример использования:
//
//	if user.HasGroup(1) {
//	    // Пользователь - администратор
//	}
func (u *User) HasGroup(groupID int) bool {
	for _, gid := range u.Groups {
		if gid == groupID {
			return true
		}
	}
	return false
}

// IsAdmin проверяет, является ли пользователь администратором.
//
// Администратор - это пользователь, принадлежащий к группе с ID=1.
// Также проверяется, что группа активна (эта проверка будет в сервисе/репозитории).
//
// Возвращает:
//   - bool: true если пользователь администратор, false иначе
//
// Пример использования:
//
//	if !user.IsAdmin() {
//	    return errors.ForbiddenError("Admin access required")
//	}
func (u *User) IsAdmin() bool {
	// Группа с ID=1 - это администраторы
	return u.HasGroup(1)
}

// HasAnyGroup проверяет, принадлежит ли пользователь хотя бы к одной из указанных групп.
//
// Параметры:
//   - groupIDs: список ID групп для проверки
//
// Возвращает:
//   - bool: true если пользователь принадлежит хотя бы к одной группе, false иначе
//
// Пример использования:
//
//	if user.HasAnyGroup([]int{1, 2}) {
//	    // Пользователь - администратор или модератор
//	}
func (u *User) HasAnyGroup(groupIDs []int) bool {
	for _, gid := range groupIDs {
		if u.HasGroup(gid) {
			return true
		}
	}
	return false
}

// HasAllGroups проверяет, принадлежит ли пользователь ко всем указанным группам.
//
// Параметры:
//   - groupIDs: список ID групп для проверки
//
// Возвращает:
//   - bool: true если пользователь принадлежит ко всем группам, false иначе
//
// Пример использования:
//
//	if user.HasAllGroups([]int{1, 2}) {
//	    // Пользователь - администратор И модератор
//	}
func (u *User) HasAllGroups(groupIDs []int) bool {
	if len(groupIDs) == 0 {
		return false
	}

	for _, gid := range groupIDs {
		if !u.HasGroup(gid) {
			return false
		}
	}
	return true
}

// SetGroups устанавливает список групп пользователя.
//
// Параметры:
//   - groupIDs: список ID групп
//
// Пример использования:
//
//	user.SetGroups([]int{1, 2})
func (u *User) SetGroups(groupIDs []int) {
	u.Groups = groupIDs
}

// AddGroup добавляет пользователя в группу (если ещё не состоит).
//
// Параметры:
//   - groupID: ID группы для добавления
//
// Возвращает:
//   - bool: true если группа была добавлена, false если пользователь уже в группе
//
// Пример использования:
//
//	added := user.AddGroup(2)
//	if added {
//	    // Группа добавлена
//	}
func (u *User) AddGroup(groupID int) bool {
	if u.HasGroup(groupID) {
		return false
	}
	u.Groups = append(u.Groups, groupID)
	return true
}

// RemoveGroup удаляет пользователя из группы.
//
// Параметры:
//   - groupID: ID группы для удаления
//
// Возвращает:
//   - bool: true если группа была удалена, false если пользователь не состоял в группе
//
// Пример использования:
//
//	removed := user.RemoveGroup(2)
//	if removed {
//	    // Группа удалена
//	}
func (u *User) RemoveGroup(groupID int) bool {
	for i, gid := range u.Groups {
		if gid == groupID {
			// Удаляем элемент из слайса
			u.Groups = append(u.Groups[:i], u.Groups[i+1:]...)
			return true
		}
	}
	return false
}

// ClearGroups удаляет все группы пользователя.
//
// Пример использования:
//
//	user.ClearGroups()
func (u *User) ClearGroups() {
	u.Groups = []int{}
}

// GetGroupsCount возвращает количество групп, к которым принадлежит пользователь.
//
// Возвращает:
//   - int: количество групп
//
// Пример использования:
//
//	count := user.GetGroupsCount()
func (u *User) GetGroupsCount() int {
	return len(u.Groups)
}
