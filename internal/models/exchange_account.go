// Package models содержит структуры данных для работы с биржевыми аккаунтами.
package models

import "time"

// ExchangeAccount представляет аккаунт на конкретной бирже.
//
// Поля соответствуют таблице EXCHANGE_ACCOUNTS.
//   - ID: уникальный идентификатор записи
//   - ExID: ID биржи, к которой относится аккаунт
//   - UID: ID пользователя-владельца аккаунта
//   - AccountName: имя аккаунта
//   - Priority: приоритет аккаунта (используется в PHP для сортировки)
//   - Active: активен ли аккаунт (true = enable, false = disable/blocked)
//   - ApiKey: публичный ключ API
//   - SecretKey: секретный ключ API (может быть пустым)
//   - AddKey: дополнительный ключ/пароль (может быть пустым)
//   - Note: произвольная заметка (может быть пустой)
//   - Deleted: soft-delete флаг
//   - DateCreate: дата создания
//   - DateModify: дата последнего изменения (nullable)
//   - UserCreated: ID пользователя, создавшего запись (nullable)
//   - UserModify: ID пользователя, изменившего запись (nullable)
type ExchangeAccount struct {
	ID          int        `json:"id" db:"ID"`
	ExID        int        `json:"exid" db:"EXID"`
	UID         int        `json:"uid" db:"UID"`
	AccountName string     `json:"account_name" db:"ACCOUNT_NAME"`
	Priority    int        `json:"priority" db:"PRIORITY"`
	Active      bool       `json:"active" db:"ACTIVE"`
	ApiKey      string     `json:"api_key" db:"API_KEY"`
	SecretKey   string     `json:"secret_key" db:"SECRET_KEY"`
	AddKey      *string    `json:"add_key,omitempty" db:"ADD_KEY"`
	Note        *string    `json:"note,omitempty" db:"NOTE"`
	Deleted     bool       `json:"deleted" db:"DELETED"`
	DateCreate  time.Time  `json:"date_create" db:"TIMESTAMP_X"`
	DateModify  *time.Time `json:"date_modify,omitempty" db:"DATE_MODIFY"`
	UserCreated *int       `json:"user_created,omitempty" db:"USER_CREATED"`
	UserModify  *int       `json:"user_modify,omitempty" db:"USER_MODIFY"`
}

// IsActive возвращает true, если аккаунт активен и не удалён.
func (a *ExchangeAccount) IsActive() bool {
	return a.Active && !a.Deleted
}
