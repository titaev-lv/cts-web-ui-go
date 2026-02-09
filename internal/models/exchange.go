// Package models - модель Exchange для работы с биржами.
// Этот файл содержит модель Exchange, которая используется для управления биржами в системе.
package models

import (
	"time"
)

// Exchange представляет биржу криптовалют.
//
// Структура соответствует таблице EXCHANGE в базе данных MySQL.
// Биржи используются для подключения к различным криптовалютным биржам через API.
//
// Поля:
//   - ID: уникальный идентификатор биржи
//   - Name: название биржи (уникальное)
//   - URL: URL биржи (например, "https://api.binance.com")
//   - BaseURL: базовый URL биржи (может отличаться от URL)
//   - ClassToFactory: имя класса для фабрики (например, "Binance", "KuCoin")
//   - Active: активна ли биржа (true = активна, false = отключена)
//   - Deleted: удалена ли биржа (soft delete, true = удалена)
//   - Description: описание биржи (может быть nil)
//   - UserCreated: ID пользователя, создавшего биржу (может быть nil)
//   - UserModify: ID пользователя, изменившего биржу (может быть nil)
//   - DateCreate: дата создания записи
//   - DateModify: дата последнего изменения (может быть nil)
//
// Пример использования:
//
//	exchange := models.Exchange{
//	    Name:        "Binance",
//	    URL:         "https://api.binance.com",
//	    BaseURL:     "https://api.binance.com",
//	    ClassToFactory: "Binance",
//	    Active:      true,
//	    Deleted:     false,
//	}
type Exchange struct {
	ID             int        `json:"id" db:"ID"`                                 // Уникальный идентификатор
	Name           string     `json:"name" db:"NAME"`                             // Название биржи (уникальное)
	URL            string     `json:"url" db:"URL"`                               // URL биржи
	BaseURL        string     `json:"base_url" db:"BASE_URL"`                     // Базовый URL биржи
	WebsocketURL   *string    `json:"websocket_url,omitempty" db:"WEBSOCKET_URL"` // WebSocket URL (может быть nil)
	ClassToFactory string     `json:"class" db:"CLASS_TO_FACTORY"`                // Имя класса для фабрики
	Active         bool       `json:"active" db:"ACTIVE"`                         // Активна ли биржа (true = активна)
	Deleted        bool       `json:"deleted" db:"DELETED"`                       // Удалена ли биржа (soft delete)
	Description    *string    `json:"description,omitempty" db:"DESCRIPTION"`     // Описание биржи (может быть nil)
	UserCreated    *int       `json:"user_created,omitempty" db:"USER_CREATED"`   // ID создателя (может быть nil)
	UserModify     *int       `json:"user_modify,omitempty" db:"USER_MODIFY"`     // ID изменившего (может быть nil)
	DateCreate     time.Time  `json:"date_create" db:"DATE_CREATE"`               // Дата создания
	DateModify     *time.Time `json:"date_modify,omitempty" db:"DATE_MODIFY"`     // Дата изменения (может быть nil)
}

// IsActive проверяет, активна ли биржа.
//
// Возвращает:
//   - bool: true если биржа активна и не удалена, false иначе
//
// Пример использования:
//
//	if !exchange.IsActive() {
//	    return errors.NotFoundError("exchange", exchangeID)
//	}
func (e *Exchange) IsActive() bool {
	return e.Active && !e.Deleted
}

// IsDeleted проверяет, удалена ли биржа (soft delete).
//
// Возвращает:
//   - bool: true если биржа удалена, false иначе
//
// Пример использования:
//
//	if exchange.IsDeleted() {
//	    return errors.NotFoundError("exchange", exchangeID)
//	}
func (e *Exchange) IsDeleted() bool {
	return e.Deleted
}
