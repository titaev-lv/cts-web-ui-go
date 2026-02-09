package repositories

import (
	"ctweb/internal/db"
	"ctweb/internal/models"
	"database/sql"
	"fmt"
)

// ExchangeAccountRepository - репозиторий для работы с аккаунтами бирж.
// Методы ориентированы на логику PHP (EXCHANGE_ACCOUNTS), включая soft-delete через поле DELETED.
type ExchangeAccountRepository struct{}

// NewExchangeAccountRepository создаёт новый экземпляр ExchangeAccountRepository.
func NewExchangeAccountRepository() *ExchangeAccountRepository {
	return &ExchangeAccountRepository{}
}

// FindByID находит аккаунт по ID и UID владельца (только не удалённые).
//
// Используем фильтр по UID, поскольку в PHP все операции выполняются от имени авторизованного пользователя.
func (r *ExchangeAccountRepository) FindByID(id int, userID int) (*models.ExchangeAccount, error) {
	query := `SELECT
		ID,
		EXID,
		UID,
		ACCOUNT_NAME,
		PRIORITY,
		ACTIVE,
		API_KEY,
		SECRET_KEY,
		ADD_KEY,
		NOTE,
		DELETED,
		TIMESTAMP_X
	FROM EXCHANGE_ACCOUNTS
	WHERE ID = ? AND UID = ? AND DELETED = 0`

	var acc models.ExchangeAccount
	var addKey, note sql.NullString

	err := db.DB.QueryRow(query, id, userID).Scan(
		&acc.ID,
		&acc.ExID,
		&acc.UID,
		&acc.AccountName,
		&acc.Priority,
		&acc.Active,
		&acc.ApiKey,
		&acc.SecretKey,
		&addKey,
		&note,
		&acc.Deleted,
		&acc.DateCreate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("exchange account with ID %d not found", id)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if addKey.Valid {
		acc.AddKey = &addKey.String
	}
	if note.Valid {
		acc.Note = &note.String
	}

	return &acc, nil
}

// FindAllByUser находит все аккаунты пользователя (не удалённые).
func (r *ExchangeAccountRepository) FindAllByUser(userID int) ([]*models.ExchangeAccount, error) {
	query := `SELECT
		ID,
		EXID,
		UID,
		ACCOUNT_NAME,
		PRIORITY,
		ACTIVE,
		API_KEY,
		SECRET_KEY,
		ADD_KEY,
		NOTE,
		DELETED,
		TIMESTAMP_X
	FROM EXCHANGE_ACCOUNTS
	WHERE UID = ? AND DELETED = 0
	ORDER BY PRIORITY DESC, EXID ASC, ID ASC`

	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var accounts []*models.ExchangeAccount
	for rows.Next() {
		var acc models.ExchangeAccount
		var addKey, note sql.NullString

		err := rows.Scan(
			&acc.ID,
			&acc.ExID,
			&acc.UID,
			&acc.AccountName,
			&acc.Priority,
			&acc.Active,
			&acc.ApiKey,
			&acc.SecretKey,
			&addKey,
			&note,
			&acc.Deleted,
			&acc.DateCreate,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		if addKey.Valid {
			acc.AddKey = &addKey.String
		}
		if note.Valid {
			acc.Note = &note.String
		}

		accounts = append(accounts, &acc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return accounts, nil
}

// CountByUser возвращает количество аккаунтов пользователя (не удалённых).
func (r *ExchangeAccountRepository) CountByUser(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM EXCHANGE_ACCOUNTS WHERE UID = ? AND DELETED = 0`

	var count int
	if err := db.DB.QueryRow(query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("database error: %w", err)
	}
	return count, nil
}

// ExistsByName проверяет уникальность имени аккаунта в рамках пользователя и биржи.
func (r *ExchangeAccountRepository) ExistsByName(userID, exchangeID int, name string) (bool, error) {
	query := `SELECT COUNT(*) FROM EXCHANGE_ACCOUNTS WHERE UID = ? AND EXID = ? AND ACCOUNT_NAME = ? AND DELETED = 0`

	var count int
	if err := db.DB.QueryRow(query, userID, exchangeID, name).Scan(&count); err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}
	return count > 0, nil
}

// ExistsByNameExcludingID проверяет уникальность имени аккаунта, исключая текущий ID.
func (r *ExchangeAccountRepository) ExistsByNameExcludingID(userID, exchangeID, excludeID int, name string) (bool, error) {
	query := `SELECT COUNT(*) FROM EXCHANGE_ACCOUNTS 
		WHERE UID = ? AND EXID = ? AND ACCOUNT_NAME = ? AND ID != ? AND DELETED = 0`

	var count int
	if err := db.DB.QueryRow(query, userID, exchangeID, name, excludeID).Scan(&count); err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}
	return count > 0, nil
}

// Create создаёт новый аккаунт (soft-delete = 0, активность по статусу).
func (r *ExchangeAccountRepository) Create(acc *models.ExchangeAccount) (int, error) {
	tx, err := db.BeginTransaction()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer db.RollbackTransaction(tx)

	query := `INSERT INTO EXCHANGE_ACCOUNTS
		(ACCOUNT_NAME, EXID, ACTIVE, UID, PRIORITY, API_KEY, SECRET_KEY, NOTE, ADD_KEY, DELETED)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`

	var note, addKey interface{}
	if acc.Note != nil {
		note = *acc.Note
	} else {
		note = nil
	}
	if acc.AddKey != nil {
		addKey = *acc.AddKey
	} else {
		addKey = nil
	}

	result, err := tx.Exec(query,
		acc.AccountName,
		acc.ExID,
		acc.Active,
		acc.UID,
		acc.Priority,
		acc.ApiKey,
		acc.SecretKey,
		note,
		addKey,
	)
	if err != nil {
		return 0, fmt.Errorf("database error: %w", err)
	}

	id, err := db.GetLastInsertID(result)
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	if err := db.CommitTransaction(tx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return int(id), nil
}

// Update обновляет аккаунт. SecretKey/AddKey обновляются только если переданы непустые значения.
func (r *ExchangeAccountRepository) Update(acc *models.ExchangeAccount) error {
	tx, err := db.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer db.RollbackTransaction(tx)

	var note interface{}
	if acc.Note != nil {
		note = *acc.Note
	} else {
		note = nil
	}

	// Базовые поля
	query := `UPDATE EXCHANGE_ACCOUNTS SET
		ACCOUNT_NAME = ?,
		EXID = ?,
		ACTIVE = ?,
		PRIORITY = ?,
		API_KEY = ?,
		NOTE = ?`
	args := []interface{}{
		acc.AccountName,
		acc.ExID,
		acc.Active,
		acc.Priority,
		acc.ApiKey,
		note,
	}

	// Опциональные поля
	if acc.SecretKey != "" {
		query += `, SECRET_KEY = ?`
		args = append(args, acc.SecretKey)
	}
	if acc.AddKey != nil && *acc.AddKey != "" {
		query += `, ADD_KEY = ?`
		args = append(args, *acc.AddKey)
	}

	query += ` WHERE ID = ? AND UID = ? AND DELETED = 0`
	args = append(args, acc.ID, acc.UID)

	result, err := tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("exchange account with ID %d not found or already deleted", acc.ID)
	}

	if err := db.CommitTransaction(tx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// SoftDelete помечает аккаунт как удалённый (DELETED = 1).
func (r *ExchangeAccountRepository) SoftDelete(id int, userID int) error {
	tx, err := db.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer db.RollbackTransaction(tx)

	query := `UPDATE EXCHANGE_ACCOUNTS SET DELETED = 1 WHERE ID = ? AND UID = ? AND DELETED = 0`
	result, err := tx.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("exchange account with ID %d not found or already deleted", id)
	}

	if err := db.CommitTransaction(tx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
