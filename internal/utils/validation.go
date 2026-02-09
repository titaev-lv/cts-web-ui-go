// Package utils содержит вспомогательные функции для валидации данных.
package utils

import (
	"ctweb/internal/repositories"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// ValidateEmail проверяет, является ли строка валидным email адресом.
//
// Параметры:
//   - email: email адрес для проверки
//
// Возвращает:
//   - error: nil, если email валиден, или описание ошибки
//
// Использует простое регулярное выражение (как в PHP: /@.+\./).
// Для более строгой проверки можно использовать более сложное регулярное выражение.
//
// Пример использования:
//
//	err := utils.ValidateEmail("user@example.com")
//	if err != nil {
//	    // email невалиден
//	}
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	// Простая проверка (как в PHP: /@.+\./)
	// Проверяем, что есть @ и после него есть точка
	// В raw строке (обратные кавычки) нужно использовать один обратный слэш для экранирования
	emailRegex := regexp.MustCompile(`@.+\.`)
	if !emailRegex.MatchString(email) {
		return errors.New("email failed")
	}

	// Дополнительная проверка: доменная часть должна содержать только допустимые символы
	// Разбиваем email на части по @
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return errors.New("email failed")
	}

	domain := parts[1]
	// Проверяем, что домен содержит только латинские буквы, цифры, точки и дефисы
	// Это предотвращает использование кириллицы и других недопустимых символов
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	if !domainRegex.MatchString(domain) {
		return errors.New("email failed")
	}

	// Проверяем, что домен не начинается и не заканчивается точкой или дефисом
	if strings.HasPrefix(domain, ".") || strings.HasPrefix(domain, "-") ||
		strings.HasSuffix(domain, ".") || strings.HasSuffix(domain, "-") {
		return errors.New("email failed")
	}

	return nil
}

// ValidateStatus проверяет, является ли статус валидным.
//
// Параметры:
//   - status: статус для проверки ("enable" или "disable")
//
// Возвращает:
//   - error: nil, если статус валиден, или описание ошибки
//
// Пример использования:
//
//	err := utils.ValidateStatus("enable")
//	if err != nil {
//	    // статус невалиден
//	}
func ValidateStatus(status string) error {
	if status != "enable" && status != "disable" {
		return errors.New("status failed")
	}
	return nil
}

// StatusToBool конвертирует строковый статус в bool.
//
// Параметры:
//   - status: статус ("enable" или "disable")
//
// Возвращает:
//   - bool: true если "enable", false если "disable"
//
// Пример использования:
//
//	active := utils.StatusToBool("enable") // true
//	active := utils.StatusToBool("disable") // false
func StatusToBool(status string) bool {
	return status == "enable"
}

// BoolToStatus конвертирует bool в строковый статус.
//
// Параметры:
//   - active: активен ли (true = "enable", false = "disable")
//
// Возвращает:
//   - string: "enable" или "disable"
//
// Пример использования:
//
//	status := utils.BoolToStatus(true) // "enable"
//	status := utils.BoolToStatus(false) // "disable"
func BoolToStatus(active bool) string {
	if active {
		return "enable"
	}
	return "disable"
}

// ParseGroupIDs парсит строку с ID групп (через запятую) в массив int.
//
// Параметры:
//   - groupsStr: строка с ID групп через запятую (например, "1,2,3")
//
// Возвращает:
//   - []int: массив ID групп
//   - error: ошибка, если не удалось распарсить
//
// Пример использования:
//
//	groupIDs, err := utils.ParseGroupIDs("1,2,3")
//	// Результат: []int{1, 2, 3}
func ParseGroupIDs(groupsStr string) ([]int, error) {
	if groupsStr == "" {
		return []int{}, errors.New("groups is required")
	}

	// Разбиваем строку по запятой
	parts := strings.Split(groupsStr, ",")
	groupIDs := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		id, err := strconv.Atoi(part)
		if err != nil {
			return nil, errors.New("invalid group ID: " + part)
		}

		groupIDs = append(groupIDs, id)
	}

	if len(groupIDs) == 0 {
		return nil, errors.New("at least one group is required")
	}

	return groupIDs, nil
}

// ValidateGroupIDs проверяет, что все ID групп существуют в базе данных.
//
// Параметры:
//   - groupIDs: массив ID групп для проверки
//
// Возвращает:
//   - error: nil, если все группы существуют, или описание ошибки
//
// Пример использования:
//
//	err := utils.ValidateGroupIDs([]int{1, 2, 3})
//	if err != nil {
//	    // не все группы существуют
//	}
func ValidateGroupIDs(groupIDs []int) error {
	if len(groupIDs) == 0 {
		return errors.New("at least one group is required")
	}

	// Импортируем репозиторий для проверки
	// ВАЖНО: Это создаёт зависимость от repositories, но это допустимо для валидации
	// В будущем можно вынести в сервис
	groupRepo := repositories.NewGroupRepository()

	for _, groupID := range groupIDs {
		_, err := groupRepo.FindByID(groupID)
		if err != nil {
			return errors.New("group with ID " + strconv.Itoa(groupID) + " not found")
		}
	}

	return nil
}

