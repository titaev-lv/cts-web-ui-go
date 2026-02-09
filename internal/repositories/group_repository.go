package repositories

import (
	"ctweb/internal/db"
	"ctweb/internal/models"
	"database/sql"
	"fmt"
	"strings"
)

// GroupRepository - репозиторий для работы с группами пользователей.
// Содержит методы для выполнения SQL запросов к таблице GROUP.
type GroupRepository struct{}

// NewGroupRepository создаёт новый экземпляр GroupRepository.
//
// Возвращает:
//   - *GroupRepository: новый репозиторий
func NewGroupRepository() *GroupRepository {
	return &GroupRepository{}
}

// FindByID находит группу по ID.
//
// Параметры:
//   - id: ID группы
//
// Возвращает:
//   - *models.Group: найденная группа
//   - error: ошибка, если группа не найдена или произошла ошибка БД
func (r *GroupRepository) FindByID(id int) (*models.Group, error) {
	query := `SELECT 
		ID, 
		NAME, 
		DESCRIPTION, 
		ACTIVE, 
		DATE_CREATE, 
		DATE_MODIFY, 
		USER_CREATED, 
		USER_MODIFY 
	FROM ` + "`GROUP`" + ` 
	WHERE ID = ?`

	var group models.Group
	var dateModify sql.NullTime
	var userCreated, userModify sql.NullInt64

	err := db.DB.QueryRow(query, id).Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&group.Active,
		&group.DateCreate,
		&dateModify,
		&userCreated,
		&userModify,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("group with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to find group by ID: %w", err)
	}

	// Обрабатываем nullable поля
	if dateModify.Valid {
		group.DateModify = &dateModify.Time
	}
	if userCreated.Valid {
		val := int(userCreated.Int64)
		group.UserCreated = &val
	}
	if userModify.Valid {
		val := int(userModify.Int64)
		group.UserModify = &val
	}

	return &group, nil
}

// FindAll находит все группы (активные и неактивные).
//
// Возвращает:
//   - []*models.Group: список всех групп
//   - error: ошибка, если произошла ошибка БД
//
// Пример использования:
//
//	repo := repositories.NewGroupRepository()
//	groups, err := repo.FindAll()
func (r *GroupRepository) FindAll() ([]*models.Group, error) {
	query := `SELECT 
		ID, 
		NAME, 
		DESCRIPTION, 
		ACTIVE, 
		DATE_CREATE, 
		DATE_MODIFY, 
		USER_CREATED, 
		USER_MODIFY 
	FROM ` + "`GROUP`" + ` 
	ORDER BY NAME ASC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var groups []*models.Group
	for rows.Next() {
		var group models.Group
		var dateModify sql.NullTime
		var userCreated, userModify sql.NullInt64

		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&group.Active,
			&group.DateCreate,
			&dateModify,
			&userCreated,
			&userModify,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// Обрабатываем nullable поля
		if dateModify.Valid {
			group.DateModify = &dateModify.Time
		}
		if userCreated.Valid {
			val := int(userCreated.Int64)
			group.UserCreated = &val
		}
		if userModify.Valid {
			val := int(userModify.Int64)
			group.UserModify = &val
		}

		groups = append(groups, &group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return groups, nil
}

// FindAllActive находит все активные группы.
//
// Возвращает:
//   - []*models.Group: список активных групп
//   - error: ошибка, если произошла ошибка БД
//
// Пример использования:
//
//	repo := repositories.NewGroupRepository()
//	activeGroups, err := repo.FindAllActive()
func (r *GroupRepository) FindAllActive() ([]*models.Group, error) {
	query := `SELECT 
		ID, 
		NAME, 
		DESCRIPTION, 
		ACTIVE, 
		DATE_CREATE, 
		DATE_MODIFY, 
		USER_CREATED, 
		USER_MODIFY 
	FROM ` + "`GROUP`" + ` 
	WHERE ACTIVE = 1
	ORDER BY NAME ASC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var groups []*models.Group
	for rows.Next() {
		var group models.Group
		var dateModify sql.NullTime
		var userCreated, userModify sql.NullInt64

		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&group.Active,
			&group.DateCreate,
			&dateModify,
			&userCreated,
			&userModify,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// Обрабатываем nullable поля
		if dateModify.Valid {
			group.DateModify = &dateModify.Time
		}
		if userCreated.Valid {
			val := int(userCreated.Int64)
			group.UserCreated = &val
		}
		if userModify.Valid {
			val := int(userModify.Int64)
			group.UserModify = &val
		}

		groups = append(groups, &group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return groups, nil
}

// FindByName находит группу по имени.
//
// Параметры:
//   - name: название группы
//
// Возвращает:
//   - *models.Group: найденная группа
//   - error: ошибка, если группа не найдена или произошла ошибка БД
//
// Используется для проверки уникальности имени группы.
//
// Пример использования:
//
//	repo := repositories.NewGroupRepository()
//	group, err := repo.FindByName("Administrators")
func (r *GroupRepository) FindByName(name string) (*models.Group, error) {
	query := `SELECT 
		ID, 
		NAME, 
		DESCRIPTION, 
		ACTIVE, 
		DATE_CREATE, 
		DATE_MODIFY, 
		USER_CREATED, 
		USER_MODIFY 
	FROM ` + "`GROUP`" + ` 
	WHERE NAME = ?`

	var group models.Group
	var dateModify sql.NullTime
	var userCreated, userModify sql.NullInt64

	err := db.DB.QueryRow(query, name).Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&group.Active,
		&group.DateCreate,
		&dateModify,
		&userCreated,
		&userModify,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("group with name '%s' not found", name)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Обрабатываем nullable поля
	if dateModify.Valid {
		group.DateModify = &dateModify.Time
	}
	if userCreated.Valid {
		val := int(userCreated.Int64)
		group.UserCreated = &val
	}
	if userModify.Valid {
		val := int(userModify.Int64)
		group.UserModify = &val
	}

	return &group, nil
}

// Count возвращает общее количество групп в базе данных.
//
// Возвращает:
//   - int: количество групп
//   - error: ошибка, если произошла ошибка БД
//
// Пример использования:
//
//	repo := repositories.NewGroupRepository()
//	count, err := repo.Count()
func (r *GroupRepository) Count() (int, error) {
	query := `SELECT COUNT(*) FROM ` + "`GROUP`" + ``

	var count int
	err := db.DB.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("database error: %w", err)
	}

	return count, nil
}

// Create создаёт новую группу в базе данных.
//
// Параметры:
//   - group: данные группы для создания
//   - userCreatedID: ID пользователя, создающего группу
//
// Возвращает:
//   - int: ID созданной группы
//   - error: ошибка, если не удалось создать группу
//
// Пример использования:
//
//	repo := repositories.NewGroupRepository()
//	group := &models.Group{
//	    Name:        "Moderators",
//	    Description: "Moderator group",
//	    Active:      true,
//	}
//	groupID, err := repo.Create(group, currentUserID)
func (r *GroupRepository) Create(group *models.Group, userCreatedID int) (int, error) {
	// Начинаем транзакцию для атомарности операции
	tx, err := db.BeginTransaction()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer db.RollbackTransaction(tx) // Откатим, если что-то пойдёт не так

	query := `INSERT INTO ` + "`GROUP`" + ` 
		(NAME, DESCRIPTION, ACTIVE, USER_CREATED, DATE_CREATE) 
		VALUES (?, ?, ?, ?, NOW())`

	result, err := tx.Exec(query,
		group.Name,
		group.Description,
		group.Active,
		userCreatedID,
	)
	if err != nil {
		return 0, fmt.Errorf("database error: %w", err)
	}

	// Получаем ID созданной группы
	groupID, err := db.GetLastInsertID(result)
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Подтверждаем транзакцию
	if err := db.CommitTransaction(tx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return int(groupID), nil
}

// Update обновляет существующую группу в базе данных.
//
// Параметры:
//   - group: данные группы для обновления (должен содержать ID)
//   - userModifyID: ID пользователя, изменяющего группу
//
// Возвращает:
//   - error: ошибка, если не удалось обновить группу
//
// Пример использования:
//
//	repo := repositories.NewGroupRepository()
//	group, _ := repo.FindByID(1)
//	group.Name = "Updated Name"
//	err := repo.Update(group, currentUserID)
func (r *GroupRepository) Update(group *models.Group, userModifyID int) error {
	// Начинаем транзакцию
	tx, err := db.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer db.RollbackTransaction(tx)

	query := `UPDATE ` + "`GROUP`" + ` SET 
		NAME = ?, 
		DESCRIPTION = ?, 
		ACTIVE = ?, 
		USER_MODIFY = ?, 
		DATE_MODIFY = NOW() 
	WHERE ID = ?`

	result, err := tx.Exec(query,
		group.Name,
		group.Description,
		group.Active,
		userModifyID,
		group.ID,
	)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Проверяем, что была обновлена хотя бы одна строка
	rowsAffected, err := db.GetRowsAffected(result)
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("group with ID %d not found", group.ID)
	}

	// Подтверждаем транзакцию
	if err := db.CommitTransaction(tx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DataTablesRequest представляет параметры запроса от DataTables.
//
// Используется для пагинации, сортировки и фильтрации данных.
type DataTablesRequest struct {
	Start  int    // Начальная позиция (offset)
	Length int    // Количество записей (limit)
	Search string // Поисковый запрос (для фильтрации)
	Order  []struct {
		Column int    // Индекс колонки для сортировки
		Dir     string // Направление сортировки ("asc" или "desc")
	}
	Columns []struct {
		Data       string // Имя поля
		Searchable bool   // Можно ли искать по этому полю
		Search     struct {
			Value string // Значение для поиска
		}
	}
}

// DataTablesResponse представляет ответ для DataTables.
type DataTablesResponse struct {
	Data            []*models.Group `json:"data"`              // Данные групп
	RecordsTotal    int             `json:"recordsTotal"`     // Общее количество записей
	RecordsFiltered int             `json:"recordsFiltered"`  // Количество записей после фильтрации
}

// FindAllWithPagination находит группы с поддержкой пагинации, сортировки и фильтрации для DataTables.
//
// Параметры:
//   - req: параметры запроса от DataTables
//
// Возвращает:
//   - *DataTablesResponse: ответ с данными и метаинформацией
//   - error: ошибка, если произошла ошибка БД
//
// Пример использования:
//
//	repo := repositories.NewGroupRepository()
//	req := &DataTablesRequest{
//	    Start:  0,
//	    Length: 10,
//	    Search: "admin",
//	    Order: []struct{Column int; Dir string}{{Column: 0, Dir: "asc"}},
//	}
//	response, err := repo.FindAllWithPagination(req)
func (r *GroupRepository) FindAllWithPagination(req *DataTablesRequest) (*DataTablesResponse, error) {
	// Маппинг колонок DataTables на поля БД
	columnMap := map[string]string{
		"id":          "ID",
		"name":        "NAME",
		"status":      "ACTIVE",
		"description": "DESCRIPTION",
	}

	// ============================================
	// ШАГ 1: Построение WHERE условий для фильтрации
	// ============================================
	var whereConditions []string
	var whereArgs []interface{}

	// Глобальный поиск (по всем колонкам)
	if req.Search != "" {
		whereConditions = append(whereConditions,
			`(g.NAME LIKE CONCAT('%', ?, '%') OR g.DESCRIPTION LIKE CONCAT('%', ?, '%'))`)
		whereArgs = append(whereArgs, req.Search, req.Search)
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
		searchValueUpper := strings.ToUpper(searchValueTrimmed)

		// Проверяем, не запрашивается ли фильтрация по NULL
		// Для DESCRIPTION может быть NULL
		if searchValueUpper == "NULL" && dbColumn == "DESCRIPTION" {
			// Фильтрация по NULL значению
			whereConditions = append(whereConditions, `g.DESCRIPTION IS NULL`)
			// Не добавляем аргументы для IS NULL
			continue
		}

		// Специальная обработка для статуса (ACTIVE)
		// В PHP используется stristr (регистронезависимый поиск подстроки)
		// Проверяем, содержит ли значение поиска "ACTIVE" или "BLOCKED" (регистронезависимо)
		// В DataTables отображается "Active" или "Blocked" (с заглавной буквы)
		if dbColumn == "ACTIVE" {
			var activeValue int
			searchValueLower := strings.ToLower(searchValueTrimmed)
			
			// Проверяем различные варианты ввода
			if strings.Contains(searchValueLower, "active") || searchValueTrimmed == "1" || strings.Contains(searchValueLower, "enable") {
				activeValue = 1
			} else if strings.Contains(searchValueLower, "blocked") || searchValueTrimmed == "0" || strings.Contains(searchValueLower, "disable") {
				activeValue = 0
			} else {
				// Если значение не распознано, пропускаем (не добавляем условие WHERE)
				continue
			}
			whereConditions = append(whereConditions, `g.ACTIVE = ?`)
			whereArgs = append(whereArgs, activeValue)
		} else {
			// Для остальных полей используем LIKE
			// Используем searchValueTrimmed вместо searchValue для консистентности
			whereConditions = append(whereConditions, `g.`+dbColumn+` LIKE CONCAT('%', ?, '%')`)
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
	// ШАГ 2: Подсчёт общего количества записей (без фильтрации)
	// ============================================
	var recordsTotal int
	countQuery := `SELECT COUNT(*) FROM ` + "`GROUP`" + ` g`
	err := db.DB.QueryRow(countQuery).Scan(&recordsTotal)
	if err != nil {
		return nil, fmt.Errorf("failed to count total records: %w", err)
	}

	// ============================================
	// ШАГ 3: Подсчёт отфильтрованных записей
	// ============================================
	var recordsFiltered int
	filteredCountQuery := `SELECT COUNT(*) FROM ` + "`GROUP`" + ` g ` + whereClause
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
	// 4. Строим ORDER BY g.ACTIVE ASC/DESC
	orderClause := "ORDER BY g.NAME ASC" // По умолчанию сортировка по имени
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
				orderClause = "ORDER BY g." + dbColumn + " " + dir
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
	query := `SELECT 
		g.ID, 
		g.NAME, 
		g.DESCRIPTION, 
		g.ACTIVE, 
		g.DATE_CREATE, 
		g.DATE_MODIFY, 
		g.USER_CREATED, 
		g.USER_MODIFY 
	FROM ` + "`GROUP`" + ` g 
	` + whereClause + ` 
	` + orderClause + ` 
	` + limitClause

	rows, err := db.DB.Query(query, whereArgs...)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var groups []*models.Group
	for rows.Next() {
		var group models.Group
		var dateModify sql.NullTime
		var userCreated, userModify sql.NullInt64

		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&group.Active,
			&group.DateCreate,
			&dateModify,
			&userCreated,
			&userModify,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// Обрабатываем nullable поля
		if dateModify.Valid {
			group.DateModify = &dateModify.Time
		}
		if userCreated.Valid {
			val := int(userCreated.Int64)
			group.UserCreated = &val
		}
		if userModify.Valid {
			val := int(userModify.Int64)
			group.UserModify = &val
		}

		groups = append(groups, &group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &DataTablesResponse{
		Data:            groups,
		RecordsTotal:    recordsTotal,
		RecordsFiltered: recordsFiltered,
	}, nil
}

// ExistsByName проверяет, существует ли группа с указанным именем.
//
// Параметры:
//   - name: название группы для проверки
//
// Возвращает:
//   - bool: true если группа с таким именем существует, false иначе
//   - error: ошибка, если произошла ошибка БД
//
// Используется для валидации уникальности имени при создании группы.
//
// Пример использования:
//
//	repo := repositories.NewGroupRepository()
//	exists, err := repo.ExistsByName("Administrators")
//	if exists {
//	    return errors.ValidationError("Group name already exists")
//	}
func (r *GroupRepository) ExistsByName(name string) (bool, error) {
	query := `SELECT COUNT(*) FROM ` + "`GROUP`" + ` WHERE NAME = ?`

	var count int
	err := db.DB.QueryRow(query, name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}

	return count > 0, nil
}

// ExistsByNameExcludingID проверяет, существует ли группа с указанным именем,
// исключая группу с указанным ID.
//
// Параметры:
//   - name: название группы для проверки
//   - excludeID: ID группы, которую нужно исключить из проверки
//
// Возвращает:
//   - bool: true если группа с таким именем существует (кроме excludeID), false иначе
//   - error: ошибка, если произошла ошибка БД
//
// Используется для валидации уникальности имени при обновлении группы.
//
// Пример использования:
//
//	repo := repositories.NewGroupRepository()
//	exists, err := repo.ExistsByNameExcludingID("Administrators", 1)
//	if exists {
//	    return errors.ValidationError("Group name already exists")
//	}
func (r *GroupRepository) ExistsByNameExcludingID(name string, excludeID int) (bool, error) {
	query := `SELECT COUNT(*) FROM ` + "`GROUP`" + ` WHERE NAME = ? AND ID != ?`

	var count int
	err := db.DB.QueryRow(query, name, excludeID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}

	return count > 0, nil
}

