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
		ea.ID,
		ea.EXID,
		e.NAME AS EXCHANGE_NAME,
		e.ACTIVE AS EXCHANGE_ACTIVE,
		ea.UID,
		ea.ACCOUNT_NAME,
		ea.PRIORITY,
		ea.ACTIVE,
		ea.API_KEY_ENC,
		ea.SECRET_KEY_ENC,
		ea.ADD_KEY_ENC,
		ea.DEK_ENC,
		ea.ENC_KEY_VERSION,
		ea.ENC_ALG,
		ea.NOTE,
		ea.DELETED,
		ea.TIMESTAMP_X
	FROM EXCHANGE_ACCOUNTS ea
	LEFT JOIN EXCHANGE e ON e.ID = ea.EXID
	WHERE ea.ID = ? AND ea.UID = ? AND ea.DELETED = 0`

	var acc models.ExchangeAccount
	var exchangeName sql.NullString
	var exchangeActive sql.NullBool
	var apiKey, secretKey, addKey, dekEnc, encAlg, note sql.NullString
	var encKeyVersion sql.NullInt64

	err := db.DB.QueryRow(query, id, userID).Scan(
		&acc.ID,
		&acc.ExID,
		&exchangeName,
		&exchangeActive,
		&acc.UID,
		&acc.AccountName,
		&acc.Priority,
		&acc.Active,
		&apiKey,
		&secretKey,
		&addKey,
		&dekEnc,
		&encKeyVersion,
		&encAlg,
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

	if apiKey.Valid {
		acc.ApiKey = apiKey.String
	}
	if exchangeName.Valid {
		acc.ExchangeName = exchangeName.String
	}
	if exchangeActive.Valid {
		acc.ExchangeActive = exchangeActive.Bool
	}
	if secretKey.Valid {
		acc.SecretKey = secretKey.String
	}
	if dekEnc.Valid {
		acc.DekEnc = dekEnc.String
	}
	if encKeyVersion.Valid {
		acc.EncKeyVersion = int(encKeyVersion.Int64)
	}
	if encAlg.Valid {
		acc.EncAlg = encAlg.String
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
		ea.ID,
		ea.EXID,
		e.NAME AS EXCHANGE_NAME,
		e.ACTIVE AS EXCHANGE_ACTIVE,
		ea.UID,
		ea.ACCOUNT_NAME,
		ea.PRIORITY,
		ea.ACTIVE,
		ea.API_KEY_ENC,
		ea.SECRET_KEY_ENC,
		ea.ADD_KEY_ENC,
		ea.DEK_ENC,
		ea.ENC_KEY_VERSION,
		ea.ENC_ALG,
		ea.NOTE,
		ea.DELETED,
		ea.TIMESTAMP_X
	FROM EXCHANGE_ACCOUNTS ea
	LEFT JOIN EXCHANGE e ON e.ID = ea.EXID
	WHERE ea.UID = ? AND ea.DELETED = 0
	ORDER BY ea.PRIORITY DESC, ea.EXID ASC, ea.ID ASC`

	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var accounts []*models.ExchangeAccount
	for rows.Next() {
		var acc models.ExchangeAccount
		var exchangeName sql.NullString
		var exchangeActive sql.NullBool
		var apiKey, secretKey, addKey, dekEnc, encAlg, note sql.NullString
		var encKeyVersion sql.NullInt64

		err := rows.Scan(
			&acc.ID,
			&acc.ExID,
			&exchangeName,
			&exchangeActive,
			&acc.UID,
			&acc.AccountName,
			&acc.Priority,
			&acc.Active,
			&apiKey,
			&secretKey,
			&addKey,
			&dekEnc,
			&encKeyVersion,
			&encAlg,
			&note,
			&acc.Deleted,
			&acc.DateCreate,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		if exchangeName.Valid {
			acc.ExchangeName = exchangeName.String
		}
		if exchangeActive.Valid {
			acc.ExchangeActive = exchangeActive.Bool
		}
		if apiKey.Valid {
			acc.ApiKey = apiKey.String
		}
		if secretKey.Valid {
			acc.SecretKey = secretKey.String
		}
		if dekEnc.Valid {
			acc.DekEnc = dekEnc.String
		}
		if encKeyVersion.Valid {
			acc.EncKeyVersion = int(encKeyVersion.Int64)
		}
		if encAlg.Valid {
			acc.EncAlg = encAlg.String
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

// FindActiveByUserExchange returns active, non-deleted user accounts for one exchange.
func (r *ExchangeAccountRepository) FindActiveByUserExchange(userID, exchangeID int) ([]*models.ExchangeAccount, error) {
	query := `SELECT
		ea.ID,
		ea.EXID,
		e.NAME AS EXCHANGE_NAME,
		e.ACTIVE AS EXCHANGE_ACTIVE,
		ea.UID,
		ea.ACCOUNT_NAME,
		ea.PRIORITY,
		ea.ACTIVE,
		ea.API_KEY_ENC,
		ea.SECRET_KEY_ENC,
		ea.ADD_KEY_ENC,
		ea.DEK_ENC,
		ea.ENC_KEY_VERSION,
		ea.ENC_ALG,
		ea.NOTE,
		ea.DELETED,
		ea.TIMESTAMP_X
	FROM EXCHANGE_ACCOUNTS ea
	JOIN EXCHANGE e ON e.ID = ea.EXID
	WHERE ea.UID = ?
	  AND ea.EXID = ?
	  AND ea.DELETED = 0
	  AND ea.ACTIVE = 1
	  AND e.DELETED = 0
	  AND e.ACTIVE = 1
	ORDER BY ea.PRIORITY DESC, ea.ID ASC`

	rows, err := db.DB.Query(query, userID, exchangeID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	accounts := make([]*models.ExchangeAccount, 0)
	for rows.Next() {
		var acc models.ExchangeAccount
		var exchangeName sql.NullString
		var exchangeActive sql.NullBool
		var apiKey, secretKey, addKey, dekEnc, encAlg, note sql.NullString
		var encKeyVersion sql.NullInt64

		err := rows.Scan(
			&acc.ID,
			&acc.ExID,
			&exchangeName,
			&exchangeActive,
			&acc.UID,
			&acc.AccountName,
			&acc.Priority,
			&acc.Active,
			&apiKey,
			&secretKey,
			&addKey,
			&dekEnc,
			&encKeyVersion,
			&encAlg,
			&note,
			&acc.Deleted,
			&acc.DateCreate,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		if exchangeName.Valid {
			acc.ExchangeName = exchangeName.String
		}
		if exchangeActive.Valid {
			acc.ExchangeActive = exchangeActive.Bool
		}
		if apiKey.Valid {
			acc.ApiKey = apiKey.String
		}
		if secretKey.Valid {
			acc.SecretKey = secretKey.String
		}
		if dekEnc.Valid {
			acc.DekEnc = dekEnc.String
		}
		if encKeyVersion.Valid {
			acc.EncKeyVersion = int(encKeyVersion.Int64)
		}
		if encAlg.Valid {
			acc.EncAlg = encAlg.String
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
		(ACCOUNT_NAME, EXID, ACTIVE, UID, PRIORITY, API_KEY_ENC, SECRET_KEY_ENC, NOTE, ADD_KEY_ENC, DEK_ENC, ENC_KEY_VERSION, ENC_ALG, DELETED)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`

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
		acc.DekEnc,
		acc.EncKeyVersion,
		acc.EncAlg,
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
		NOTE = ?`
	args := []interface{}{
		acc.AccountName,
		acc.ExID,
		acc.Active,
		acc.Priority,
		note,
	}

	// Опциональные поля
	if acc.ApiKey != "" {
		query += `, API_KEY_ENC = ?`
		args = append(args, acc.ApiKey)
	}
	if acc.SecretKey != "" {
		query += `, SECRET_KEY_ENC = ?`
		args = append(args, acc.SecretKey)
	}
	if acc.AddKey != nil && *acc.AddKey != "" {
		query += `, ADD_KEY_ENC = ?`
		args = append(args, *acc.AddKey)
	}
	if acc.DekEnc != "" {
		query += `, DEK_ENC = ?, ENC_KEY_VERSION = ?, ENC_ALG = ?`
		args = append(args, acc.DekEnc, acc.EncKeyVersion, acc.EncAlg)
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
		// MySQL may report 0 affected rows when values are unchanged.
		// Verify record still exists before treating this as not found.
		var exists int
		err = tx.QueryRow(`SELECT 1 FROM EXCHANGE_ACCOUNTS WHERE ID = ? AND UID = ? AND DELETED = 0`, acc.ID, acc.UID).Scan(&exists)
		if err == nil {
			if commitErr := db.CommitTransaction(tx); commitErr != nil {
				return fmt.Errorf("failed to commit transaction: %w", commitErr)
			}
			return nil
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("failed to verify exchange account existence: %w", err)
		}
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
