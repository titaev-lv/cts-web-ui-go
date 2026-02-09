// Package models - модель Group для работы с группами пользователей.
// Этот файл содержит базовую модель Group, которая будет расширена в Фазе 3.
package models

import (
	"time"
)

// Group представляет группу пользователей.
//
// Структура соответствует таблице GROUP в базе данных MySQL.
// Группы используются для управления правами доступа пользователей.
//
// Поля:
//   - ID: уникальный идентификатор группы
//   - Name: название группы (уникальное)
//   - Description: описание группы (может быть пустым)
//   - Active: активна ли группа (true = активна, false = отключена)
//   - DateCreate: дата создания записи
//   - DateModify: дата последнего изменения (может быть nil)
//   - UserCreated: ID пользователя, создавшего группу (может быть nil)
//   - UserModify: ID пользователя, изменившего группу (может быть nil)
//
// Специальные группы:
//   - ID=1: Администраторы (Admin)
//   - ID=2: Пользователи (User) - обычные пользователи
//
// Пример использования:
//
//	group := models.Group{
//	    Name:        "Administrators",
//	    Description: "System administrators",
//	    Active:      true,
//	}
type Group struct {
	ID          int        `json:"id" db:"ID"`                    // Уникальный идентификатор
	Name        string     `json:"name" db:"NAME"`                // Название группы (уникальное)
	Description string     `json:"description" db:"DESCRIPTION"`   // Описание группы
	Active      bool       `json:"active" db:"ACTIVE"`           // Активна ли группа (true = активна)
	DateCreate  time.Time  `json:"date_create" db:"DATE_CREATE"` // Дата создания
	DateModify  *time.Time `json:"date_modify,omitempty" db:"DATE_MODIFY"` // Дата изменения (может быть nil)
	UserCreated *int       `json:"user_created,omitempty" db:"USER_CREATED"` // ID создателя (может быть nil)
	UserModify  *int       `json:"user_modify,omitempty" db:"USER_MODIFY"`   // ID изменившего (может быть nil)
}

// IsActive проверяет, активна ли группа.
//
// Возвращает:
//   - bool: true если группа активна, false если отключена
//
// Пример использования:
//
//	if !group.IsActive() {
//	    return errors.NotFoundError("group", groupID)
//	}
func (g *Group) IsActive() bool {
	return g.Active
}

// IsAdminGroup проверяет, является ли группа группой администраторов.
//
// Группа администраторов имеет ID=1.
//
// Возвращает:
//   - bool: true если это группа администраторов, false иначе
//
// Пример использования:
//
//	if group.IsAdminGroup() {
//	    // Это группа администраторов
//	}
func (g *Group) IsAdminGroup() bool {
	return g.ID == 1
}

