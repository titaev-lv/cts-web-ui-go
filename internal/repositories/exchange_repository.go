package repositories

import (
	"ctweb/internal/db"
	"ctweb/internal/models"
	"database/sql"
	"fmt"
	"strings"
)

// ExchangeDataTablesRow представляет строку данных для DataTables.
// Используется для форматирования ответа в формате, ожидаемом DataTables.
type ExchangeDataTablesRow struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Active         string  `json:"status"` // "Active" или "Blocked"
	URL            string  `json:"url"`
	BaseURL        string  `json:"base_url"`
	WebsocketURL   *string `json:"websocket_url"`
	ClassToFactory string  `json:"class"`
}

// ExchangeDataTablesResponse представляет ответ для DataTables для бирж.
type ExchangeDataTablesResponse struct {
	Data            []*ExchangeDataTablesRow `json:"data"`            // Данные бирж
	RecordsTotal    int                      `json:"recordsTotal"`    // Общее количество записей
	RecordsFiltered int                      `json:"recordsFiltered"` // Количество записей после фильтрации
}

// ExchangeRepository - репозиторий для работы с биржами.
// Содержит методы для выполнения SQL запросов к таблице EXCHANGE.
type ExchangeRepository struct{}

// NewExchangeRepository создаёт новый экземпляр ExchangeRepository.
//
// Возвращает:
//   - *ExchangeRepository: новый репозиторий
func NewExchangeRepository() *ExchangeRepository {
	return &ExchangeRepository{}
}

// FindByID находит биржу по ID (только не удалённые).
//
// Параметры:
//   - id: ID биржи
//
// Возвращает:
//   - *models.Exchange: найденная биржа
//   - error: ошибка, если биржа не найдена или произошла ошибка БД
func (r *ExchangeRepository) FindByID(id int) (*models.Exchange, error) {
	query := `SELECT 
		ID, 
		NAME, 
		URL, 
		BASE_URL, 
		WEBSOCKET_URL,
		CLASS_TO_FACTORY, 
		ACTIVE, 
		DELETED, 
		DESCRIPTION, 
		USER_CREATED, 
		USER_MODIFY, 
		DATE_CREATE, 
		DATE_MODIFY 
	FROM EXCHANGE 
	WHERE ID = ? AND DELETED = 0`

	var exchange models.Exchange
	var dateModify sql.NullTime
	var userCreated, userModify sql.NullInt64
	var description sql.NullString
	var websocketURL sql.NullString

	err := db.DB.QueryRow(query, id).Scan(
		&exchange.ID,
		&exchange.Name,
		&exchange.URL,
		&exchange.BaseURL,
		&websocketURL,
		&exchange.ClassToFactory,
		&exchange.Active,
		&exchange.Deleted,
		&description,
		&userCreated,
		&userModify,
		&exchange.DateCreate,
		&dateModify,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("exchange with ID %d not found", id)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Обрабатываем nullable поля
	if description.Valid {
		exchange.Description = &description.String
	}
	if websocketURL.Valid {
		exchange.WebsocketURL = &websocketURL.String
	}
	if dateModify.Valid {
		exchange.DateModify = &dateModify.Time
	}
	if userCreated.Valid {
		val := int(userCreated.Int64)
		exchange.UserCreated = &val
	}
	if userModify.Valid {
		val := int(userModify.Int64)
		exchange.UserModify = &val
	}

	return &exchange, nil
}

// FindAll находит все биржи (активные и неактивные, но не удалённые).
//
// Возвращает:
//   - []*models.Exchange: список всех бирж
//   - error: ошибка, если произошла ошибка БД
func (r *ExchangeRepository) FindAll() ([]*models.Exchange, error) {
	query := `SELECT 
		ID, 
		NAME, 
		URL, 
		BASE_URL, 
		CLASS_TO_FACTORY, 
		ACTIVE, 
		DELETED, 
		DESCRIPTION, 
		USER_CREATED, 
		USER_MODIFY, 
		DATE_CREATE, 
		DATE_MODIFY 
	FROM EXCHANGE 
	WHERE DELETED = 0 
	ORDER BY ID ASC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var exchanges []*models.Exchange
	for rows.Next() {
		var exchange models.Exchange
		var dateModify sql.NullTime
		var userCreated, userModify sql.NullInt64
		var description sql.NullString

		err := rows.Scan(
			&exchange.ID,
			&exchange.Name,
			&exchange.URL,
			&exchange.BaseURL,
			&exchange.ClassToFactory,
			&exchange.Active,
			&exchange.Deleted,
			&description,
			&userCreated,
			&userModify,
			&exchange.DateCreate,
			&dateModify,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// Обрабатываем nullable поля
		if description.Valid {
			exchange.Description = &description.String
		}
		if dateModify.Valid {
			exchange.DateModify = &dateModify.Time
		}
		if userCreated.Valid {
			val := int(userCreated.Int64)
			exchange.UserCreated = &val
		}
		if userModify.Valid {
			val := int(userModify.Int64)
			exchange.UserModify = &val
		}

		exchanges = append(exchanges, &exchange)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return exchanges, nil
}

// FindAllActive находит все активные биржи (не удалённые).
//
// Возвращает:
//   - []*models.Exchange: список активных бирж
//   - error: ошибка, если произошла ошибка БД
func (r *ExchangeRepository) FindAllActive() ([]*models.Exchange, error) {
	query := `SELECT 
		ID, 
		NAME, 
		URL, 
		BASE_URL, 
		CLASS_TO_FACTORY, 
		ACTIVE, 
		DELETED, 
		DESCRIPTION, 
		USER_CREATED, 
		USER_MODIFY, 
		DATE_CREATE, 
		DATE_MODIFY 
	FROM EXCHANGE 
	WHERE ACTIVE = 1 AND DELETED = 0 
	ORDER BY ID ASC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var exchanges []*models.Exchange
	for rows.Next() {
		var exchange models.Exchange
		var dateModify sql.NullTime
		var userCreated, userModify sql.NullInt64
		var description sql.NullString

		err := rows.Scan(
			&exchange.ID,
			&exchange.Name,
			&exchange.URL,
			&exchange.BaseURL,
			&exchange.ClassToFactory,
			&exchange.Active,
			&exchange.Deleted,
			&description,
			&userCreated,
			&userModify,
			&exchange.DateCreate,
			&dateModify,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// Обрабатываем nullable поля
		if description.Valid {
			exchange.Description = &description.String
		}
		if dateModify.Valid {
			exchange.DateModify = &dateModify.Time
		}
		if userCreated.Valid {
			val := int(userCreated.Int64)
			exchange.UserCreated = &val
		}
		if userModify.Valid {
			val := int(userModify.Int64)
			exchange.UserModify = &val
		}

		exchanges = append(exchanges, &exchange)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return exchanges, nil
}

// FindByName находит биржу по имени (только не удалённые).
//
// Параметры:
//   - name: название биржи
//
// Возвращает:
//   - *models.Exchange: найденная биржа
//   - error: ошибка, если биржа не найдена или произошла ошибка БД
//
// Используется для проверки уникальности имени биржи.
func (r *ExchangeRepository) FindByName(name string) (*models.Exchange, error) {
	query := `SELECT 
		ID, 
		NAME, 
		URL, 
		BASE_URL, 
		CLASS_TO_FACTORY, 
		ACTIVE, 
		DELETED, 
		DESCRIPTION, 
		USER_CREATED, 
		USER_MODIFY, 
		DATE_CREATE, 
		DATE_MODIFY 
	FROM EXCHANGE 
	WHERE NAME = ? AND DELETED = 0`

	var exchange models.Exchange
	var dateModify sql.NullTime
	var userCreated, userModify sql.NullInt64
	var description sql.NullString

	err := db.DB.QueryRow(query, name).Scan(
		&exchange.ID,
		&exchange.Name,
		&exchange.URL,
		&exchange.BaseURL,
		&exchange.ClassToFactory,
		&exchange.Active,
		&exchange.Deleted,
		&description,
		&userCreated,
		&userModify,
		&exchange.DateCreate,
		&dateModify,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("exchange with name '%s' not found", name)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Обрабатываем nullable поля
	if description.Valid {
		exchange.Description = &description.String
	}
	if dateModify.Valid {
		exchange.DateModify = &dateModify.Time
	}
	if userCreated.Valid {
		val := int(userCreated.Int64)
		exchange.UserCreated = &val
	}
	if userModify.Valid {
		val := int(userModify.Int64)
		exchange.UserModify = &val
	}

	return &exchange, nil
}

// Count возвращает общее количество бирж в базе данных (не удалённых).
//
// Возвращает:
//   - int: количество бирж
//   - error: ошибка, если произошла ошибка БД
func (r *ExchangeRepository) Count() (int, error) {
	query := `SELECT COUNT(*) FROM EXCHANGE WHERE DELETED = 0`

	var count int
	err := db.DB.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("database error: %w", err)
	}

	return count, nil
}

// Create создаёт новую биржу в базе данных.
//
// Параметры:
//   - exchange: данные биржи для создания
//   - userCreatedID: ID пользователя, создающего биржу
//
// Возвращает:
//   - int: ID созданной биржи
//   - error: ошибка, если не удалось создать биржу
func (r *ExchangeRepository) Create(exchange *models.Exchange, userCreatedID int) (int, error) {
	// Начинаем транзакцию для атомарности операции
	tx, err := db.BeginTransaction()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer db.RollbackTransaction(tx) // Откатим, если что-то пойдёт не так

	query := `INSERT INTO EXCHANGE 
		(NAME, URL, ACTIVE, USER_CREATED, BASE_URL, CLASS_TO_FACTORY, DESCRIPTION, DELETED, DATE_CREATE) 
		VALUES (?, ?, ?, ?, ?, ?, ?, 0, NOW())`

	var descriptionValue interface{}
	if exchange.Description != nil {
		descriptionValue = *exchange.Description
	} else {
		descriptionValue = nil
	}

	result, err := tx.Exec(query,
		exchange.Name,
		exchange.URL,
		exchange.Active,
		userCreatedID,
		exchange.BaseURL,
		exchange.ClassToFactory,
		descriptionValue,
	)
	if err != nil {
		return 0, fmt.Errorf("database error: %w", err)
	}

	// Получаем ID созданной биржи
	exchangeID, err := db.GetLastInsertID(result)
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Подтверждаем транзакцию
	if err := db.CommitTransaction(tx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return int(exchangeID), nil
}

// Update обновляет существующую биржу в базе данных.
//
// Параметры:
//   - exchange: данные биржи для обновления (должен содержать ID)
//   - userModifyID: ID пользователя, изменяющего биржу
//
// Возвращает:
//   - error: ошибка, если не удалось обновить биржу
func (r *ExchangeRepository) Update(exchange *models.Exchange, userModifyID int) error {
	// Начинаем транзакцию
	tx, err := db.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer db.RollbackTransaction(tx)

	query := `UPDATE EXCHANGE SET 
		NAME = ?, 
		DESCRIPTION = ?, 
		ACTIVE = ?, 
		USER_MODIFY = ?, 
		DATE_MODIFY = NOW(), 
		URL = ?, 
		BASE_URL = ?, 
		CLASS_TO_FACTORY = ? 
	WHERE ID = ? AND DELETED = 0`

	var descriptionValue interface{}
	if exchange.Description != nil {
		descriptionValue = *exchange.Description
	} else {
		descriptionValue = nil
	}

	result, err := tx.Exec(query,
		exchange.Name,
		descriptionValue,
		exchange.Active,
		userModifyID,
		exchange.URL,
		exchange.BaseURL,
		exchange.ClassToFactory,
		exchange.ID,
	)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Проверяем, что была обновлена хотя бы одна строка
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("exchange with ID %d not found or already deleted", exchange.ID)
	}

	// Подтверждаем транзакцию
	if err := db.CommitTransaction(tx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ExistsByName проверяет, существует ли биржа с указанным именем (не удалённая).
//
// Параметры:
//   - name: название биржи
//
// Возвращает:
//   - bool: true если биржа существует, false иначе
//   - error: ошибка, если произошла ошибка БД
func (r *ExchangeRepository) ExistsByName(name string) (bool, error) {
	query := `SELECT COUNT(*) FROM EXCHANGE WHERE NAME = ? AND DELETED = 0`

	var count int
	err := db.DB.QueryRow(query, name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}

	return count > 0, nil
}

// ExistsByNameExcludingID проверяет, существует ли биржа с указанным именем,
// исключая биржу с указанным ID (не удалённые).
//
// Параметры:
//   - name: название биржи
//   - excludeID: ID биржи, которую нужно исключить из проверки
//
// Возвращает:
//   - bool: true если биржа существует, false иначе
//   - error: ошибка, если произошла ошибка БД
//
// Используется при обновлении биржи для проверки уникальности имени.
func (r *ExchangeRepository) ExistsByNameExcludingID(name string, excludeID int) (bool, error) {
	query := `SELECT COUNT(*) FROM EXCHANGE WHERE NAME = ? AND ID != ? AND DELETED = 0`

	var count int
	err := db.DB.QueryRow(query, name, excludeID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}

	return count > 0, nil
}

// FindAllWithPagination находит биржи с поддержкой пагинации, сортировки и фильтрации для DataTables.
//
// Параметры:
//   - req: параметры запроса от DataTables (используется DataTablesRequest из GroupRepository)
//
// Возвращает:
//   - *ExchangeDataTablesResponse: ответ с данными и метаинформацией
//   - error: ошибка, если произошла ошибка БД
//
// Пример использования:
//
//	repo := repositories.NewExchangeRepository()
//	req := &DataTablesRequest{
//	    Start:  0,
//	    Length: 10,
//	    Search: "binance",
//	    Order: []struct{Column int; Dir string}{{Column: 0, Dir: "asc"}},
//	}
//	response, err := repo.FindAllWithPagination(req)
func (r *ExchangeRepository) FindAllWithPagination(req *DataTablesRequest) (*ExchangeDataTablesResponse, error) {
	// Маппинг колонок DataTables на поля БД
	columnMap := map[string]string{
		"id":            "ID",
		"name":          "NAME",
		"status":        "ACTIVE",
		"url":           "URL",
		"base_url":      "BASE_URL",
		"websocket_url": "WEBSOCKET_URL",
		"class":         "CLASS_TO_FACTORY",
	}

	// ============================================
	// ШАГ 1: Построение WHERE условий для фильтрации
	// ============================================
	var whereConditions []string
	var whereArgs []interface{}

	// Всегда фильтруем по DELETED = 0 (soft delete)
	whereConditions = append(whereConditions, "ch.DELETED = 0")

	// Глобальный поиск (по всем колонкам)
	if req.Search != "" {
		whereConditions = append(whereConditions,
			`(ch.NAME LIKE CONCAT('%', ?, '%') OR ch.URL LIKE CONCAT('%', ?, '%') OR ch.BASE_URL LIKE CONCAT('%', ?, '%') OR ch.CLASS_TO_FACTORY LIKE CONCAT('%', ?, '%'))`)
		whereArgs = append(whereArgs, req.Search, req.Search, req.Search, req.Search)
	}

	// Поиск по отдельным колонкам
	for _, col := range req.Columns {
		if !col.Searchable {
			continue
		}

		// Проверяем, есть ли значение для поиска
		if col.Search.Value == "" {
			continue
		}

		dbColumn, exists := columnMap[col.Data]
		if !exists {
			continue
		}

		searchValue := col.Search.Value
		searchValueTrimmed := strings.TrimSpace(searchValue)
		searchValueLower := strings.ToLower(searchValueTrimmed)

		// Специальная обработка для статуса (ACTIVE)
		// В PHP используется stristr (регистронезависимый поиск подстроки)
		// Проверяем, содержит ли значение поиска "ACTIVE" или "BLOCKED" (регистронезависимо)
		// В DataTables отображается "Active" или "Blocked" (с заглавной буквы)
		if dbColumn == "ACTIVE" {
			var activeValue int

			// Проверяем различные варианты ввода
			if strings.Contains(searchValueLower, "active") || searchValueTrimmed == "1" || strings.Contains(searchValueLower, "enable") {
				activeValue = 1
			} else if strings.Contains(searchValueLower, "blocked") || searchValueTrimmed == "0" || strings.Contains(searchValueLower, "disable") {
				activeValue = 0
			} else {
				// Если значение не распознано, пропускаем (не добавляем условие WHERE)
				continue
			}
			whereConditions = append(whereConditions, `ch.ACTIVE = ?`)
			whereArgs = append(whereArgs, activeValue)
		} else {
			// Для остальных полей используем LIKE
			// Используем searchValueTrimmed вместо searchValue для консистентности
			whereConditions = append(whereConditions, `ch.`+dbColumn+` LIKE CONCAT('%', ?, '%')`)
			whereArgs = append(whereArgs, searchValueTrimmed)
		}
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + whereConditions[0]
		for i := 1; i < len(whereConditions); i++ {
			whereClause += " AND " + whereConditions[i]
		}
	}

	// ============================================
	// ШАГ 2: Подсчёт общего количества записей (без фильтрации, но с DELETED = 0)
	// ============================================
	var recordsTotal int
	countQuery := `SELECT COUNT(*) FROM EXCHANGE ch WHERE ch.DELETED = 0`
	err := db.DB.QueryRow(countQuery).Scan(&recordsTotal)
	if err != nil {
		return nil, fmt.Errorf("failed to count total records: %w", err)
	}

	// ============================================
	// ШАГ 3: Подсчёт отфильтрованных записей
	// ============================================
	var recordsFiltered int
	filteredCountQuery := `SELECT COUNT(*) FROM EXCHANGE ch ` + whereClause
	err = db.DB.QueryRow(filteredCountQuery, whereArgs...).Scan(&recordsFiltered)
	if err != nil {
		return nil, fmt.Errorf("failed to count filtered records: %w", err)
	}

	// ============================================
	// ШАГ 4: Построение ORDER BY
	// ============================================
	// Логика построения ORDER BY:
	// 1. Берем индекс колонки из req.Order[0].Column (например, 3)
	// 2. Получаем имя колонки DataTables из req.Columns[3].Data (например, "status")
	// 3. Ищем в columnMap по ключу "status" и получаем имя колонки БД "ACTIVE"
	// 4. Строим ORDER BY ch.ACTIVE ASC/DESC
	orderClause := "ORDER BY ch.ID ASC" // По умолчанию сортировка по ID
	if len(req.Order) > 0 && len(req.Columns) > 0 {
		orderCol := req.Order[0].Column
		if orderCol >= 0 && orderCol < len(req.Columns) {
			colName := req.Columns[orderCol].Data // Имя колонки DataTables (например, "status")
			if dbColumn, exists := columnMap[colName]; exists {
				// dbColumn - имя колонки БД (например, "ACTIVE")
				dir := "ASC"
				if req.Order[0].Dir == "desc" {
					dir = "DESC"
				}
				orderClause = "ORDER BY ch." + dbColumn + " " + dir
			}
		}
	}

	// ============================================
	// ШАГ 5: Построение LIMIT
	// ============================================
	limitClause := fmt.Sprintf("LIMIT %d, %d", req.Start, req.Length)

	// ============================================
	// ШАГ 6: Выполнение основного запроса
	// ============================================
	// Используем подзапрос, как в PHP, для соответствия формату DataTables
	query := `SELECT
		COUNT(*) OVER() AS cnt,
		q.*
		FROM
			(
			SELECT
				ch.ID,
				ch.NAME,
				CASE
					WHEN ch.ACTIVE = 0 THEN 'Blocked'
					ELSE 'Active'
				END AS ACTIVE,
				ch.URL,
				ch.BASE_URL,
				ch.WEBSOCKET_URL,
				ch.CLASS_TO_FACTORY
			FROM
				EXCHANGE ch
			` + whereClause + `
			) AS q
		` + orderClause + ` 
		` + limitClause

	rows, err := db.DB.Query(query, whereArgs...)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var exchanges []*ExchangeDataTablesRow
	for rows.Next() {
		var exchange ExchangeDataTablesRow
		var cnt int // Игнорируем COUNT(*) OVER()

		err := rows.Scan(
			&cnt,
			&exchange.ID,
			&exchange.Name,
			&exchange.Active,
			&exchange.URL,
			&exchange.BaseURL,
			&exchange.WebsocketURL,
			&exchange.ClassToFactory,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		exchanges = append(exchanges, &exchange)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &ExchangeDataTablesResponse{
		Data:            exchanges,
		RecordsTotal:    recordsTotal,
		RecordsFiltered: recordsFiltered,
	}, nil
}
