package services

import (
	"ctweb/internal/db"
	"ctweb/internal/models"
	"log"
)

// FindUserByUsername находит пользователя по логину.
//
// Параметры:
//   - username: логин пользователя для поиска
//
// Возвращает:
//   - *models.User: найденный пользователь
//   - error: ошибка, если пользователь не найден или произошла ошибка БД
//
// Пример использования:
//
//	user, err := services.FindUserByUsername("admin")
//	if err != nil {
//	    // Пользователь не найден или ошибка БД
//	}
func FindUserByUsername(username string) (*models.User, error) {
	// Выполняем SQL запрос для поиска пользователя по логину
	// Используем prepared statement для защиты от SQL injection
	row := db.DB.QueryRow("SELECT ID, LOGIN, PASSWORD, DATE_CREATE FROM USER WHERE LOGIN=?", username)
	log.Println("user=", username)

	u := models.User{}
	// Сканируем результат в структуру User
	// Используем правильные имена полей: Login вместо Username, DateCreate вместо Created
	err := row.Scan(&u.ID, &u.Login, &u.Password, &u.DateCreate)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
