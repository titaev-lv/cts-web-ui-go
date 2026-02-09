// Package repositories содержит репозитории для работы с базой данных.
// Репозитории отвечают за выполнение SQL запросов и маппинг результатов в модели.
package repositories

import (
	"ctweb/internal/db"
	"ctweb/internal/logger"
	"ctweb/internal/models"
	"database/sql"
	"fmt"
	"strings"
)

// UserRepository - репозиторий для работы с пользователями.
// Содержит методы для выполнения SQL запросов к таблице USER.
type UserRepository struct{}

// NewUserRepository создаёт новый экземпляр UserRepository.
//
// Возвращает:
//   - *UserRepository: новый репозиторий
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// FindByLogin находит пользователя по логину.
//
// Параметры:
//   - login: логин пользователя
//
// Возвращает:
//   - *models.User: найденный пользователь
//   - error: ошибка, если пользователь не найден или произошла ошибка БД
//
// Пример использования:
//
//	repo := repositories.NewUserRepository()
//	user, err := repo.FindByLogin("admin")
func (r *UserRepository) FindByLogin(login string) (*models.User, error) {
	query := `SELECT 
		ID, 
		LOGIN, 
		PASSWORD, 
		EMAIL, 
		ACTIVE, 
		NAME, 
		LAST_NAME, 
		TOKEN, 
		TIMEZONE, 
		DATE_CREATE, 
		DATE_MODIFY, 
		TIMESTAMP_X, 
		USER_CREATED, 
		USER_MODIFY 
	FROM USER 
	WHERE LOGIN = ?`

	var user models.User
	var token, timezone sql.NullString
	var dateModify, timestampX sql.NullTime
	var userCreated, userModify sql.NullInt64

	err := db.DB.QueryRow(query, login).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.Email,
		&user.Active,
		&user.Name,
		&user.LastName,
		&token,
		&timezone,
		&user.DateCreate,
		&dateModify,
		&timestampX,
		&userCreated,
		&userModify,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Обрабатываем nullable поля
	if token.Valid {
		user.Token = token.String
	}
	if timezone.Valid {
		user.Timezone = timezone.String
	}
	if dateModify.Valid {
		user.DateModify = &dateModify.Time
	}
	if timestampX.Valid {
		user.TimestampX = &timestampX.Time
	}
	if userCreated.Valid {
		val := int(userCreated.Int64)
		user.UserCreated = &val
	}
	if userModify.Valid {
		val := int(userModify.Int64)
		user.UserModify = &val
	}

	return &user, nil
}

// FindByID находит пользователя по ID.
//
// Параметры:
//   - id: ID пользователя
//
// Возвращает:
//   - *models.User: найденный пользователь
//   - error: ошибка, если пользователь не найден или произошла ошибка БД
func (r *UserRepository) FindByID(id int) (*models.User, error) {
	query := `SELECT 
		ID, 
		LOGIN, 
		PASSWORD, 
		EMAIL, 
		ACTIVE, 
		NAME, 
		LAST_NAME, 
		TOKEN, 
		TIMEZONE, 
		DATE_CREATE, 
		DATE_MODIFY, 
		TIMESTAMP_X, 
		USER_CREATED, 
		USER_MODIFY 
	FROM USER 
	WHERE ID = ?`

	var user models.User
	var token, timezone sql.NullString
	var dateModify, timestampX sql.NullTime
	var userCreated, userModify sql.NullInt64

	err := db.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.Email,
		&user.Active,
		&user.Name,
		&user.LastName,
		&token,
		&timezone,
		&user.DateCreate,
		&dateModify,
		&timestampX,
		&userCreated,
		&userModify,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	// Обрабатываем nullable поля
	if token.Valid {
		user.Token = token.String
	}
	if timezone.Valid {
		user.Timezone = timezone.String
	}
	if dateModify.Valid {
		user.DateModify = &dateModify.Time
	}
	if timestampX.Valid {
		user.TimestampX = &timestampX.Time
	}
	if userCreated.Valid {
		val := int(userCreated.Int64)
		user.UserCreated = &val
	}
	if userModify.Valid {
		val := int(userModify.Int64)
		user.UserModify = &val
	}

	return &user, nil
}

// FindByLoginAndToken находит пользователя по логину и токену (для Remember Me).
//
// Параметры:
//   - login: логин пользователя
//   - token: токен из cookie
//
// Возвращает:
//   - *models.User: найденный пользователь
//   - error: ошибка, если пользователь не найден или произошла ошибка БД
func (r *UserRepository) FindByLoginAndToken(login, token string) (*models.User, error) {
	query := `SELECT 
		ID, 
		LOGIN, 
		PASSWORD, 
		EMAIL, 
		ACTIVE, 
		NAME, 
		LAST_NAME, 
		TOKEN, 
		TIMEZONE, 
		DATE_CREATE, 
		DATE_MODIFY, 
		TIMESTAMP_X, 
		USER_CREATED, 
		USER_MODIFY 
	FROM USER 
	WHERE LOGIN = ? AND TOKEN = ?`

	var user models.User
	var tokenVal, timezone sql.NullString
	var dateModify, timestampX sql.NullTime
	var userCreated, userModify sql.NullInt64

	err := db.DB.QueryRow(query, login, token).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.Email,
		&user.Active,
		&user.Name,
		&user.LastName,
		&tokenVal,
		&timezone,
		&user.DateCreate,
		&dateModify,
		&timestampX,
		&userCreated,
		&userModify,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Обрабатываем nullable поля
	if tokenVal.Valid {
		user.Token = tokenVal.String
	}
	if timezone.Valid {
		user.Timezone = timezone.String
	}
	if dateModify.Valid {
		user.DateModify = &dateModify.Time
	}
	if timestampX.Valid {
		user.TimestampX = &timestampX.Time
	}
	if userCreated.Valid {
		val := int(userCreated.Int64)
		user.UserCreated = &val
	}
	if userModify.Valid {
		val := int(userModify.Int64)
		user.UserModify = &val
	}

	return &user, nil
}

// FindGroupsByUserID находит все группы пользователя (включая неактивные).
//
// Параметры:
//   - userID: ID пользователя
//
// Возвращает:
//   - []int: список ID всех групп пользователя (включая неактивные)
//   - error: ошибка, если произошла ошибка БД
//
// ВАЖНО: Возвращаются ВСЕ группы пользователя, включая неактивные,
// так как при редактировании нужно показывать все присвоенные группы.
func (r *UserRepository) FindGroupsByUserID(userID int) ([]int, error) {
	// ВАЖНО: GROUP - это зарезервированное слово в MySQL, поэтому экранируем его обратными кавычками
	// Используем LEFT JOIN, чтобы получить все группы пользователя, даже если группа неактивна
	// Убираем фильтр g.ACTIVE = 1, чтобы вернуть все группы (как в PHP)
	query := `SELECT ug.GID 
	FROM USERS_GROUP ug
	WHERE ug.UID = ?`

	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var groups []int
	for rows.Next() {
		var gid int
		if err := rows.Scan(&gid); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		groups = append(groups, gid)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return groups, nil
}

// UpdateToken обновляет токен пользователя (для Remember Me).
//
// Параметры:
//   - userID: ID пользователя
//   - token: новый токен (пустая строка для удаления токена)
//
// Возвращает:
//   - error: ошибка, если произошла ошибка БД
func (r *UserRepository) UpdateToken(userID int, token string) error {
	query := `UPDATE USER SET TOKEN = ? WHERE ID = ?`
	_, err := db.DB.Exec(query, token, userID)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}

// UpdateTimestamp обновляет временную метку последней активности пользователя.
//
// Параметры:
//   - userID: ID пользователя
//
// Возвращает:
//   - error: ошибка, если произошла ошибка БД
func (r *UserRepository) UpdateTimestamp(userID int) error {
	query := `UPDATE USER SET TIMESTAMP_X = NOW() WHERE ID = ?`
	_, err := db.DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}

// FindAll находит всех пользователей (активных и неактивных).
//
// Возвращает:
//   - []*models.User: список всех пользователей
//   - error: ошибка, если произошла ошибка БД
//
// Пример использования:
//
//	repo := repositories.NewUserRepository()
//	users, err := repo.FindAll()
func (r *UserRepository) FindAll() ([]*models.User, error) {
	query := `SELECT 
		ID, 
		LOGIN, 
		PASSWORD, 
		EMAIL, 
		ACTIVE, 
		NAME, 
		LAST_NAME, 
		TOKEN, 
		TIMEZONE, 
		DATE_CREATE, 
		DATE_MODIFY, 
		TIMESTAMP_X, 
		USER_CREATED, 
		USER_MODIFY 
	FROM USER 
	ORDER BY LOGIN ASC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var token, timezone sql.NullString
		var dateModify, timestampX sql.NullTime
		var userCreated, userModify sql.NullInt64

		err := rows.Scan(
			&user.ID,
			&user.Login,
			&user.Password,
			&user.Email,
			&user.Active,
			&user.Name,
			&user.LastName,
			&token,
			&timezone,
			&user.DateCreate,
			&dateModify,
			&timestampX,
			&userCreated,
			&userModify,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// Обрабатываем nullable поля
		if token.Valid {
			user.Token = token.String
		}
		if timezone.Valid {
			user.Timezone = timezone.String
		}
		if dateModify.Valid {
			user.DateModify = &dateModify.Time
		}
		if timestampX.Valid {
			user.TimestampX = &timestampX.Time
		}
		if userCreated.Valid {
			val := int(userCreated.Int64)
			user.UserCreated = &val
		}
		if userModify.Valid {
			val := int(userModify.Int64)
			user.UserModify = &val
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

// FindAllActive находит всех активных пользователей.
//
// Возвращает:
//   - []*models.User: список активных пользователей
//   - error: ошибка, если произошла ошибка БД
//
// Пример использования:
//
//	repo := repositories.NewUserRepository()
//	activeUsers, err := repo.FindAllActive()
func (r *UserRepository) FindAllActive() ([]*models.User, error) {
	query := `SELECT 
		ID, 
		LOGIN, 
		PASSWORD, 
		EMAIL, 
		ACTIVE, 
		NAME, 
		LAST_NAME, 
		TOKEN, 
		TIMEZONE, 
		DATE_CREATE, 
		DATE_MODIFY, 
		TIMESTAMP_X, 
		USER_CREATED, 
		USER_MODIFY 
	FROM USER 
	WHERE ACTIVE = 1
	ORDER BY LOGIN ASC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var token, timezone sql.NullString
		var dateModify, timestampX sql.NullTime
		var userCreated, userModify sql.NullInt64

		err := rows.Scan(
			&user.ID,
			&user.Login,
			&user.Password,
			&user.Email,
			&user.Active,
			&user.Name,
			&user.LastName,
			&token,
			&timezone,
			&user.DateCreate,
			&dateModify,
			&timestampX,
			&userCreated,
			&userModify,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// Обрабатываем nullable поля
		if token.Valid {
			user.Token = token.String
		}
		if timezone.Valid {
			user.Timezone = timezone.String
		}
		if dateModify.Valid {
			user.DateModify = &dateModify.Time
		}
		if timestampX.Valid {
			user.TimestampX = &timestampX.Time
		}
		if userCreated.Valid {
			val := int(userCreated.Int64)
			user.UserCreated = &val
		}
		if userModify.Valid {
			val := int(userModify.Int64)
			user.UserModify = &val
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

// Count возвращает общее количество пользователей в базе данных.
//
// Возвращает:
//   - int: количество пользователей
//   - error: ошибка, если произошла ошибка БД
//
// Пример использования:
//
//	repo := repositories.NewUserRepository()
//	count, err := repo.Count()
func (r *UserRepository) Count() (int, error) {
	query := `SELECT COUNT(*) FROM USER`

	var count int
	err := db.DB.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("database error: %w", err)
	}

	return count, nil
}

// ExistsByLogin проверяет, существует ли пользователь с указанным логином.
//
// Параметры:
//   - login: логин для проверки
//
// Возвращает:
//   - bool: true если пользователь с таким логином существует, false иначе
//   - error: ошибка, если произошла ошибка БД
//
// Используется для валидации уникальности логина при создании пользователя.
//
// Пример использования:
//
//	repo := repositories.NewUserRepository()
//	exists, err := repo.ExistsByLogin("admin")
//	if exists {
//	    return errors.ValidationError("Login already exists")
//	}
func (r *UserRepository) ExistsByLogin(login string) (bool, error) {
	query := `SELECT COUNT(*) FROM USER WHERE LOGIN = ?`

	var count int
	err := db.DB.QueryRow(query, login).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}

	return count > 0, nil
}

// ExistsByLoginExcludingID проверяет, существует ли пользователь с указанным логином,
// исключая пользователя с указанным ID.
//
// Параметры:
//   - login: логин для проверки
//   - excludeID: ID пользователя, которого нужно исключить из проверки
//
// Возвращает:
//   - bool: true если пользователь с таким логином существует (кроме excludeID), false иначе
//   - error: ошибка, если произошла ошибка БД
//
// Используется для валидации уникальности логина при обновлении пользователя.
//
// Пример использования:
//
//	repo := repositories.NewUserRepository()
//	exists, err := repo.ExistsByLoginExcludingID("admin", 1)
//	if exists {
//	    return errors.ValidationError("Login already exists")
//	}
func (r *UserRepository) ExistsByLoginExcludingID(login string, excludeID int) (bool, error) {
	query := `SELECT COUNT(*) FROM USER WHERE LOGIN = ? AND ID != ?`

	var count int
	err := db.DB.QueryRow(query, login, excludeID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}

	return count > 0, nil
}

// ExistsByEmail проверяет, существует ли пользователь с указанным email.
//
// Параметры:
//   - email: email для проверки
//
// Возвращает:
//   - bool: true если пользователь с таким email существует, false иначе
//   - error: ошибка, если произошла ошибка БД
//
// Используется для валидации уникальности email при создании пользователя.
//
// Пример использования:
//
//	repo := repositories.NewUserRepository()
//	exists, err := repo.ExistsByEmail("user@example.com")
//	if exists {
//	    return errors.ValidationError("Email already exists")
//	}
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	query := `SELECT COUNT(*) FROM USER WHERE EMAIL = ?`

	var count int
	err := db.DB.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}

	return count > 0, nil
}

// ExistsByEmailExcludingID проверяет, существует ли пользователь с указанным email,
// исключая пользователя с указанным ID.
//
// Параметры:
//   - email: email для проверки
//   - excludeID: ID пользователя, которого нужно исключить из проверки
//
// Возвращает:
//   - bool: true если пользователь с таким email существует (кроме excludeID), false иначе
//   - error: ошибка, если произошла ошибка БД
//
// Используется для валидации уникальности email при обновлении пользователя.
//
// Пример использования:
//
//	repo := repositories.NewUserRepository()
//	exists, err := repo.ExistsByEmailExcludingID("user@example.com", 1)
//	if exists {
//	    return errors.ValidationError("Email already exists")
//	}
func (r *UserRepository) ExistsByEmailExcludingID(email string, excludeID int) (bool, error) {
	query := `SELECT COUNT(*) FROM USER WHERE EMAIL = ? AND ID != ?`

	var count int
	err := db.DB.QueryRow(query, email, excludeID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}

	return count > 0, nil
}

// SetUserGroups устанавливает группы для пользователя (удаляет старые и добавляет новые).
//
// Параметры:
//   - tx: транзакция (если nil, будет создана новая)
//   - userID: ID пользователя
//   - groupIDs: список ID групп для установки
//
// Возвращает:
//   - error: ошибка, если произошла ошибка БД
//
// ВАЖНО: Этот метод должен вызываться внутри транзакции (например, в Create или Update).
//
// Пример использования:
//
//	tx, _ := db.BeginTransaction()
//	defer db.RollbackTransaction(tx)
//	err := repo.SetUserGroups(tx, userID, []int{1, 2})
func (r *UserRepository) SetUserGroups(tx *sql.Tx, userID int, groupIDs []int) error {
	// Если транзакция не передана, создаём новую
	var err error
	if tx == nil {
		tx, err = db.BeginTransaction()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer db.RollbackTransaction(tx)
	}

	// Удаляем все существующие связи пользователя с группами
	deleteQuery := `DELETE FROM USERS_GROUP WHERE UID = ?`
	_, err = tx.Exec(deleteQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user groups: %w", err)
	}

	// Добавляем новые связи
	if len(groupIDs) > 0 {
		insertQuery := `INSERT INTO USERS_GROUP (UID, GID) VALUES (?, ?)`
		for _, groupID := range groupIDs {
			_, err = tx.Exec(insertQuery, userID, groupID)
			if err != nil {
				return fmt.Errorf("failed to insert user group: %w", err)
			}
		}
	}

	return nil
}

// Create создаёт нового пользователя в базе данных вместе с группами.
//
// Параметры:
//   - user: данные пользователя для создания (пароль должен быть уже захеширован)
//   - groupIDs: список ID групп, к которым будет принадлежать пользователь
//   - userCreatedID: ID пользователя, создающего запись
//
// Возвращает:
//   - int: ID созданного пользователя
//   - error: ошибка, если не удалось создать пользователя
//
// ВАЖНО: Использует транзакцию для атомарности операции.
// Пароль должен быть уже захеширован (bcrypt) перед вызовом этого метода.
//
// Пример использования:
//
//	repo := repositories.NewUserRepository()
//	hashedPassword, _ := utils.PasswordHash("MyPassword123")
//	user := &models.User{
//	    Login:    "newuser",
//	    Password: hashedPassword,
//	    Email:    "user@example.com",
//	    Name:     "John",
//	    LastName: "Doe",
//	    Active:   true,
//	    Timezone: "UTC",
//	}
//	userID, err := repo.Create(user, []int{2}, currentUserID)
func (r *UserRepository) Create(user *models.User, groupIDs []int, userCreatedID int) (int, error) {
	// Начинаем транзакцию
	tx, err := db.BeginTransaction()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer db.RollbackTransaction(tx) // Откатим, если что-то пойдёт не так

	// Вставляем пользователя
	query := `INSERT INTO USER 
		(LOGIN, PASSWORD, EMAIL, ACTIVE, NAME, LAST_NAME, TIMEZONE, USER_CREATED, DATE_CREATE) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())`

	result, err := tx.Exec(query,
		user.Login,
		user.Password, // Должен быть уже захеширован
		user.Email,
		user.Active,
		user.Name,
		user.LastName,
		user.Timezone,
		userCreatedID,
	)
	if err != nil {
		return 0, fmt.Errorf("database error: %w", err)
	}

	// Получаем ID созданного пользователя
	userID, err := db.GetLastInsertID(result)
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Устанавливаем группы пользователя
	if len(groupIDs) > 0 {
		err = r.SetUserGroups(tx, int(userID), groupIDs)
		if err != nil {
			return 0, fmt.Errorf("failed to set user groups: %w", err)
		}
	}

	// Подтверждаем транзакцию
	if err := db.CommitTransaction(tx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return int(userID), nil
}

// Update обновляет существующего пользователя в базе данных.
//
// Параметры:
//   - user: данные пользователя для обновления (должен содержать ID)
//   - groupIDs: список ID групп, к которым будет принадлежать пользователь (nil = не изменять)
//   - updatePassword: если true, обновляет пароль (пароль должен быть уже захеширован)
//   - userModifyID: ID пользователя, изменяющего запись
//
// Возвращает:
//   - error: ошибка, если не удалось обновить пользователя
//
// ВАЖНО: Использует транзакцию для атомарности операции.
// Если updatePassword = true, пароль должен быть уже захеширован (bcrypt).
// Если groupIDs = nil, группы не изменяются. Если groupIDs = []int{}, все группы удаляются.
//
// Пример использования:
//
//	repo := repositories.NewUserRepository()
//	user, _ := repo.FindByID(1)
//	user.Name = "Updated Name"
//	err := repo.Update(user, []int{1, 2}, false, currentUserID)
func (r *UserRepository) Update(user *models.User, groupIDs []int, updatePassword bool, userModifyID int) error {
	// Начинаем транзакцию
	tx, err := db.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer db.RollbackTransaction(tx)

	// Формируем запрос UPDATE
	var query string
	var args []interface{}

	if updatePassword {
		query = `UPDATE USER SET 
			LOGIN = ?, 
			PASSWORD = ?, 
			EMAIL = ?, 
			ACTIVE = ?, 
			NAME = ?, 
			LAST_NAME = ?, 
			TIMEZONE = ?, 
			USER_MODIFY = ?, 
			DATE_MODIFY = NOW() 
		WHERE ID = ?`
		args = []interface{}{
			user.Login,
			user.Password, // Должен быть уже захеширован
			user.Email,
			user.Active,
			user.Name,
			user.LastName,
			user.Timezone,
			userModifyID,
			user.ID,
		}
	} else {
		query = `UPDATE USER SET 
			LOGIN = ?, 
			EMAIL = ?, 
			ACTIVE = ?, 
			NAME = ?, 
			LAST_NAME = ?, 
			TIMEZONE = ?, 
			USER_MODIFY = ?, 
			DATE_MODIFY = NOW() 
		WHERE ID = ?`
		args = []interface{}{
			user.Login,
			user.Email,
			user.Active,
			user.Name,
			user.LastName,
			user.Timezone,
			userModifyID,
			user.ID,
		}
	}

	result, err := tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Проверяем, что была обновлена хотя бы одна строка
	rowsAffected, err := db.GetRowsAffected(result)
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %d not found", user.ID)
	}

	// Обновляем группы пользователя (если переданы)
	if groupIDs != nil {
		err = r.SetUserGroups(tx, user.ID, groupIDs)
		if err != nil {
			return fmt.Errorf("failed to set user groups: %w", err)
		}
	}

	// Подтверждаем транзакцию
	if err := db.CommitTransaction(tx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UserDataTablesRequest представляет параметры запроса от DataTables для пользователей.
//
// Используется для пагинации, сортировки и фильтрации данных.
type UserDataTablesRequest struct {
	Start  int    // Начальная позиция (offset)
	Length int    // Количество записей (limit)
	Search string // Поисковый запрос (для фильтрации)
	Order  []struct {
		Column int    // Индекс колонки для сортировки
		Dir    string // Направление сортировки ("asc" или "desc")
	}
	Columns []struct {
		Data       string // Имя поля
		Searchable bool   // Можно ли искать по этому полю
		Search     struct {
			Value string // Значение для поиска
		}
	}
}

// UserDataTablesResponse представляет ответ для DataTables с данными пользователей.
type UserDataTablesResponse struct {
	Data            []*UserDataTablesRow `json:"data"`            // Данные пользователей
	RecordsTotal    int                  `json:"recordsTotal"`    // Общее количество записей
	RecordsFiltered int                  `json:"recordsFiltered"` // Количество записей после фильтрации
}

// UserDataTablesRow представляет строку данных пользователя для DataTables.
// Включает информацию о пользователе и его группах.
type UserDataTablesRow struct {
	ID         int    `json:"id"`
	Login      string `json:"login"`
	Groups     string `json:"groups"` // Список групп через запятую (например, "Admin, User")
	Active     string `json:"active"` // "Active" или "Blocked"
	Name       string `json:"name"`   // Полное имя (LAST_NAME + NAME)
	Email      string `json:"email"`
	CreateDate string `json:"create_date"` // Форматированная дата
	ModifyDate string `json:"modify_date"` // Форматированная дата (может быть пустой)
	Timestamp  string `json:"timestamp"`   // Форматированная дата последней активности (может быть пустой)
}

// FindAllWithPagination находит пользователей с поддержкой пагинации, сортировки и фильтрации для DataTables.
//
// Параметры:
//   - req: параметры запроса от DataTables
//
// Возвращает:
//   - *UserDataTablesResponse: ответ с данными и метаинформацией
//   - error: ошибка, если произошла ошибка БД
//
// Пример использования:
//
//	repo := repositories.NewUserRepository()
//	req := &UserDataTablesRequest{
//	    Start:  0,
//	    Length: 10,
//	    Search: "admin",
//	    Order: []struct{Column int; Dir string}{{Column: 0, Dir: "asc"}},
//	}
//	response, err := repo.FindAllWithPagination(req)
func (r *UserRepository) FindAllWithPagination(req *UserDataTablesRequest) (*UserDataTablesResponse, error) {
	// Маппинг колонок DataTables на поля БД
	columnMap := map[string]string{
		"id":          "ID",
		"login":       "LOGIN",
		"groups":      "GROUP",
		"active":      "ACTIVE",
		"email":       "EMAIL",
		"name":        "NAME",
		"create_date": "DATE_CREATE",
		"modify_date": "DATE_MODIFY",
		"timestamp":   "TIMESTAMP_X",
	}

	// ============================================
	// ШАГ 1: Построение WHERE условий для фильтрации
	// ============================================
	var whereConditions []string
	var whereArgs []interface{}

	// Глобальный поиск (по всем колонкам)
	if req.Search != "" {
		whereConditions = append(whereConditions,
			`(u.LOGIN LIKE CONCAT('%', ?, '%') OR u.EMAIL LIKE CONCAT('%', ?, '%') OR 
			u.NAME LIKE CONCAT('%', ?, '%') OR u.LAST_NAME LIKE CONCAT('%', ?, '%'))`)
		whereArgs = append(whereArgs, req.Search, req.Search, req.Search, req.Search)
	}

	// Поиск по отдельным колонкам
	for _, col := range req.Columns {
		if !col.Searchable || col.Search.Value == "" {
			continue
		}

		dbColumn, exists := columnMap[col.Data]
		if !exists {
			continue
		}

		searchValue := col.Search.Value

		// Специальная обработка для разных полей
		switch dbColumn {
		case "DATE_CREATE", "DATE_MODIFY", "TIMESTAMP_X":
			// Обработка дат: поддержка операторов > и <, частичного совпадения и NULL
			// В подзапросе даты форматируются как DATE_FORMAT(u.DATE_CREATE, '%d-%m-%Y %H:%i:%s')
			// WHERE условия применяются к подзапросу, поэтому используем DATE_FORMAT в WHERE
			searchValueTrimmed := strings.TrimSpace(searchValue)
			searchValueUpper := strings.ToUpper(searchValueTrimmed)

			// Определяем имя поля в таблице USER
			var dateField string
			switch dbColumn {
			case "DATE_CREATE":
				dateField = "u.DATE_CREATE"
			case "DATE_MODIFY":
				dateField = "u.DATE_MODIFY"
			case "TIMESTAMP_X":
				dateField = "u.TIMESTAMP_X"
			}

			// Проверяем, не запрашивается ли фильтрация по NULL
			// Для DATE_CREATE NULL невозможен, но для DATE_MODIFY и TIMESTAMP_X - возможен
			if searchValueUpper == "NULL" && (dbColumn == "DATE_MODIFY" || dbColumn == "TIMESTAMP_X") {
				// Фильтрация по NULL значению
				if dbColumn == "TIMESTAMP_X" {
					// Для TIMESTAMP проверяем, что он NULL или пустой (как в SELECT)
					whereConditions = append(whereConditions,
						`(u.TIMESTAMP_X IS NULL OR u.TIMESTAMP_X <= u.DATE_CREATE)`)
				} else {
					// Для DATE_MODIFY просто проверяем NULL
					whereConditions = append(whereConditions, dateField+` IS NULL`)
				}
				// Не добавляем аргументы для IS NULL
				continue
			}

			// Проверяем, есть ли оператор сравнения в начале строки
			if strings.HasPrefix(searchValueTrimmed, ">") {
				// Больше чем: u.DATE_CREATE > STR_TO_DATE('значение', '%d-%m-%Y %H:%i:%s')
				// Для операторов используем сравнение с исходной датой
				dateValue := strings.TrimSpace(searchValueTrimmed[1:])
				if dateValue != "" {
					// Парсим дату в формате dd-mm-yyyy или dd-mm-yyyy HH:ii:ss
					// Если время не указано, добавляем 00:00:00
					if !strings.Contains(dateValue, ":") {
						dateValue = dateValue + " 00:00:00"
					}
					whereConditions = append(whereConditions, dateField+` > STR_TO_DATE(?, '%d-%m-%Y %H:%i:%s')`)
					whereArgs = append(whereArgs, dateValue)
				}
			} else if strings.HasPrefix(searchValueTrimmed, "<") {
				// Меньше чем: u.DATE_CREATE < STR_TO_DATE('значение', '%d-%m-%Y %H:%i:%s')
				dateValue := strings.TrimSpace(searchValueTrimmed[1:])
				if dateValue != "" {
					// Парсим дату в формате dd-mm-yyyy или dd-mm-yyyy HH:ii:ss
					// Если время не указано, добавляем 23:59:59 для включения всего дня
					if !strings.Contains(dateValue, ":") {
						dateValue = dateValue + " 23:59:59"
					}
					whereConditions = append(whereConditions, dateField+` < STR_TO_DATE(?, '%d-%m-%Y %H:%i:%s')`)
					whereArgs = append(whereArgs, dateValue)
				}
			} else {
				// Частичное совпадение: DATE_FORMAT(u.DATE_CREATE, '%d-%m-%Y %H:%i:%s') LIKE '%значение%'
				// Ищем по отформатированной дате (как в PHP)
				// Для MODIFY_DATE и TIMESTAMP учитываем NULL значения
				if dbColumn == "DATE_MODIFY" {
					// MODIFY_DATE может быть NULL, используем COALESCE
					whereConditions = append(whereConditions, `COALESCE(DATE_FORMAT(`+dateField+`, '%d-%m-%Y %H:%i:%s'), '') LIKE CONCAT('%', ?, '%')`)
				} else if dbColumn == "TIMESTAMP_X" {
					// TIMESTAMP может быть NULL или пустым, используем CASE (как в SELECT)
					whereConditions = append(whereConditions,
						`CASE 
							WHEN `+dateField+` > u.DATE_CREATE THEN COALESCE(DATE_FORMAT(`+dateField+`, '%d-%m-%Y %H:%i:%s'), '')
							ELSE ''
						END LIKE CONCAT('%', ?, '%')`)
				} else {
					// DATE_CREATE всегда заполнено
					whereConditions = append(whereConditions, `DATE_FORMAT(`+dateField+`, '%d-%m-%Y %H:%i:%s') LIKE CONCAT('%', ?, '%')`)
				}
				whereArgs = append(whereArgs, searchValueTrimmed)
			}
		case "ACTIVE":
			// В PHP используется stristr (регистронезависимый поиск подстроки)
			// Проверяем, содержит ли значение поиска "ACTIVE" или "BLOCKED" (регистронезависимо)
			// В DataTables отображается "Active" или "Blocked" (с заглавной буквы)
			var activeValue int
			searchValueTrimmed := strings.TrimSpace(searchValue)
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
			whereConditions = append(whereConditions, `u.ACTIVE = ?`)
			whereArgs = append(whereArgs, activeValue)
		case "NAME":
			// Поиск по имени и фамилии
			// Если есть пробел, разбиваем на части (имя и фамилия)
			if strings.Contains(searchValue, " ") {
				parts := strings.Fields(searchValue)
				if len(parts) >= 2 {
					whereConditions = append(whereConditions,
						`(u.NAME LIKE CONCAT('%', ?, '%') OR u.LAST_NAME LIKE CONCAT('%', ?, '%') OR 
						(u.LAST_NAME LIKE CONCAT('%', ?, '%') AND u.NAME LIKE CONCAT('%', ?, '%')))`)
					whereArgs = append(whereArgs, parts[0], parts[1], parts[0], parts[1])
				} else {
					whereConditions = append(whereConditions,
						`(u.NAME LIKE CONCAT('%', ?, '%') OR u.LAST_NAME LIKE CONCAT('%', ?, '%'))`)
					whereArgs = append(whereArgs, searchValue, searchValue)
				}
			} else {
				whereConditions = append(whereConditions,
					`(u.NAME LIKE CONCAT('%', ?, '%') OR u.LAST_NAME LIKE CONCAT('%', ?, '%'))`)
				whereArgs = append(whereArgs, searchValue, searchValue)
			}
		case "GROUP":
			// Поиск по группам (по имени группы)
			whereConditions = append(whereConditions, `g.NAME LIKE CONCAT('%', ?, '%')`)
			whereArgs = append(whereArgs, searchValue)
		default:
			// Для остальных полей используем LIKE
			whereConditions = append(whereConditions, `u.`+dbColumn+` LIKE CONCAT('%', ?, '%')`)
			whereArgs = append(whereArgs, searchValue)
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
	countQuery := `SELECT COUNT(DISTINCT u.ID) FROM USER u`
	err := db.DB.QueryRow(countQuery).Scan(&recordsTotal)
	if err != nil {
		return nil, fmt.Errorf("failed to count total records: %w", err)
	}

	// ============================================
	// ШАГ 3: Подсчёт отфильтрованных записей
	// ============================================
	var recordsFiltered int
	filteredCountQuery := `SELECT COUNT(DISTINCT u.ID) 
		FROM USER u
		RIGHT JOIN USERS_GROUP ug ON u.ID = ug.UID
		LEFT JOIN ` + "`GROUP`" + ` g ON g.ID = ug.GID ` + whereClause
	err = db.DB.QueryRow(filteredCountQuery, whereArgs...).Scan(&recordsFiltered)
	if err != nil {
		return nil, fmt.Errorf("failed to count filtered records: %w", err)
	}

	// ============================================
	// ШАГ 4: Построение ORDER BY
	// ============================================
	// Маппинг колонок DataTables на алиасы в подзапросе (q)
	// В подзапросе используются алиасы: ID, LOGIN, GROUPS, ACTIVE, NAME, EMAIL, CREATE_DATE, MODIFY_DATE, TIMESTAMP
	orderColumnMap := map[string]string{
		"id":          "ID",
		"login":       "LOGIN",
		"groups":      "GROUPS",
		"active":      "ACTIVE",
		"email":       "EMAIL",
		"name":        "NAME",
		"create_date": "CREATE_DATE",
		"modify_date": "MODIFY_DATE",
		"timestamp":   "TIMESTAMP",
	}

	orderClause := "ORDER BY q.ID ASC" // По умолчанию сортировка по ID (q - алиас подзапроса)
	if len(req.Order) > 0 && len(req.Columns) > 0 {
		orderCol := req.Order[0].Column
		if orderCol < len(req.Columns) {
			colName := req.Columns[orderCol].Data
			if orderColumn, exists := orderColumnMap[colName]; exists {
				dir := "ASC"
				if req.Order[0].Dir == "desc" {
					dir = "DESC"
				}
				// GROUPS - зарезервированное слово, нужно экранировать
				if colName == "groups" {
					orderClause = "ORDER BY q." + "`GROUPS`" + " " + dir
				} else {
					orderClause = "ORDER BY q." + orderColumn + " " + dir
				}
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
		COUNT(*) OVER() AS cnt,
		q.*
	FROM (
		SELECT
			u.ID,
			u.LOGIN,
			REPLACE(GROUP_CONCAT(g.NAME), ',', ', ') AS ` + "`GROUPS`" + `,
			CASE
				WHEN u.ACTIVE = 0 THEN 'Blocked'
				ELSE 'Active'
			END AS ACTIVE,
			CONCAT(u.LAST_NAME, ' ', u.NAME) AS NAME,
			u.EMAIL,
			DATE_FORMAT(u.DATE_CREATE, '%d-%m-%Y %H:%i:%s') AS CREATE_DATE,
			COALESCE(DATE_FORMAT(u.DATE_MODIFY, '%d-%m-%Y %H:%i:%s'), '') AS MODIFY_DATE,
			CASE
				WHEN u.TIMESTAMP_X > u.DATE_CREATE THEN COALESCE(DATE_FORMAT(u.TIMESTAMP_X, '%d-%m-%Y %H:%i:%s'), '')
				ELSE ''
			END AS TIMESTAMP
		FROM USER u
		RIGHT JOIN USERS_GROUP ug ON u.ID = ug.UID
		LEFT JOIN ` + "`GROUP`" + ` g ON g.ID = ug.GID`

	if whereClause != "" {
		// Добавляем WHERE условия в подзапрос
		query += " " + whereClause
	}

	query += ` GROUP BY u.ID
	) AS q ` + orderClause + ` ` + limitClause

	logger.Debug().
		Str("query", query).
		Interface("args", whereArgs).
		Msg("Executing FindAllWithPagination query")

	rows, err := db.DB.Query(query, whereArgs...)
	if err != nil {
		logger.Error().
			Err(err).
			Str("query", query).
			Interface("args", whereArgs).
			Msg("Database query error")
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var users []*UserDataTablesRow
	for rows.Next() {
		var user UserDataTablesRow
		var cnt int // Игнорируем COUNT(*) OVER()

		err := rows.Scan(
			&cnt,
			&user.ID,
			&user.Login,
			&user.Groups,
			&user.Active,
			&user.Name,
			&user.Email,
			&user.CreateDate,
			&user.ModifyDate,
			&user.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	logger.Debug().
		Int("users_count", len(users)).
		Msg("FindAllWithPagination completed")

	return &UserDataTablesResponse{
		Data:            users,
		RecordsTotal:    recordsTotal,
		RecordsFiltered: recordsFiltered,
	}, nil
}
