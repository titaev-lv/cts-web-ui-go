package utils

import (
	"ctweb/internal/config"
	"errors"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHash создаёт bcrypt хеш пароля.
//
// Параметры:
//   - password: пароль в открытом виде
//
// Возвращает:
//   - string: bcrypt хеш пароля
//   - error: ошибка, если не удалось создать хеш
//
// Использование:
//   hash, err := PasswordHash("myPassword123")
//   if err != nil {
//       // обработка ошибки
//   }
//
// Примечание:
//   Использует стоимость (cost) из конфигурации (по умолчанию 10).
//   Чем выше cost, тем безопаснее, но медленнее хеширование.
func PasswordHash(password string) (string, error) {
	// Получаем конфигурацию для определения стоимости хеширования
	cfg := config.Get()
	cost := cfg.Security.BcryptCost

	// Если cost не задан или равен 0, используем значение по умолчанию
	if cost <= 0 {
		cost = bcrypt.DefaultCost // Обычно это 10
	}

	// Генерируем bcrypt хеш
	// bcrypt.GenerateFromPassword автоматически добавляет соль (salt)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}

	// Возвращаем хеш как строку
	return string(hash), nil
}

// PasswordVerify проверяет, соответствует ли пароль хешу.
//
// Параметры:
//   - password: пароль в открытом виде
//   - hash: bcrypt хеш пароля (из базы данных)
//
// Возвращает:
//   - bool: true, если пароль соответствует хешу, false в противном случае
//   - error: ошибка, если произошла проблема при проверке
//
// Использование:
//   isValid, err := PasswordVerify("myPassword123", storedHash)
//   if err != nil {
//       // обработка ошибки
//   }
//   if isValid {
//       // пароль верный
//   }
//
// Примечание:
//   Использует bcrypt.CompareHashAndPassword, который безопасно сравнивает
//   пароль с хешем, защищая от timing attacks.
func PasswordVerify(password, hash string) (bool, error) {
	// Сравниваем пароль с хешем
	// bcrypt.CompareHashAndPassword возвращает nil, если пароль совпадает
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		// Если ошибка - пароль не совпадает
		// Это может быть bcrypt.ErrMismatchedHashAndPassword или другая ошибка
		return false, err
	}

	// Пароль совпадает
	return true, nil
}

// PasswordValidate проверяет, соответствует ли пароль требованиям безопасности.
//
// Требования к паролю (как в PHP):
//   - Минимум 10 символов
//   - Должен содержать строчные буквы (a-z)
//   - Должен содержать заглавные буквы (A-Z)
//   - Должен содержать цифры (0-9)
//   - Может содержать только буквы и цифры (без спецсимволов)
//
// Параметры:
//   - password: пароль для проверки
//
// Возвращает:
//   - error: nil, если пароль валиден, или описание ошибки
//
// Использование:
//   err := PasswordValidate("MyPassword123")
//   if err != nil {
//       // пароль не соответствует требованиям
//       fmt.Println(err.Error())
//   }
//
// Примеры:
//   PasswordValidate("Short1") -> error: "password must be at least 10 characters"
//   PasswordValidate("nouppercase123") -> error: "password must contain uppercase letters"
//   PasswordValidate("NOLOWERCASE123") -> error: "password must contain lowercase letters"
//   PasswordValidate("NoNumbers") -> error: "password must contain numbers"
//   PasswordValidate("ValidPass123") -> nil (валидный пароль)
func PasswordValidate(password string) error {
	// Проверка минимальной длины
	if len(password) < 10 {
		return errors.New("password must be at least 10 characters")
	}

	// Проверяем наличие различных типов символов
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasInvalid := false

	// Проходим по каждому символу пароля
	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		default:
			// Если символ не буква и не цифра - это недопустимый символ
			// Согласно PHP регулярному выражению: [a-zA-Z\d]
			hasInvalid = true
		}
	}

	// Проверяем наличие всех требуемых типов символов
	if !hasLower {
		return errors.New("password must contain lowercase letters")
	}
	if !hasUpper {
		return errors.New("password must contain uppercase letters")
	}
	if !hasDigit {
		return errors.New("password must contain numbers")
	}
	if hasInvalid {
		return errors.New("password must contain only letters and numbers")
	}

	// Пароль соответствует всем требованиям
	return nil
}

// PasswordValidateWithConfirm проверяет пароль и его подтверждение.
//
// Параметры:
//   - password: пароль
//   - passwordConfirm: подтверждение пароля
//
// Возвращает:
//   - error: nil, если пароль валиден и совпадает с подтверждением,
//            или описание ошибки
//
// Использование:
//   err := PasswordValidateWithConfirm("MyPassword123", "MyPassword123")
//   if err != nil {
//       // пароль не валиден или не совпадает
//   }
//
// Примечание:
//   Сначала проверяет, что пароли совпадают, затем валидирует пароль.
func PasswordValidateWithConfirm(password, passwordConfirm string) error {
	// Проверяем, что пароли совпадают
	if password != passwordConfirm {
		return errors.New("password confirmation does not match")
	}

	// Валидируем сам пароль
	return PasswordValidate(password)
}

// PasswordValidateRegex проверяет пароль с помощью регулярного выражения
// (как в PHP: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{10,}$/).
//
// Параметры:
//   - password: пароль для проверки
//
// Возвращает:
//   - bool: true, если пароль соответствует регулярному выражению
//   - error: ошибка, если не удалось выполнить проверку
//
// Использование:
//   isValid, err := PasswordValidateRegex("MyPassword123")
//   if err != nil {
//       // обработка ошибки
//   }
//
// Примечание:
//   Это альтернативный способ проверки, используемый для совместимости с PHP.
//   Рекомендуется использовать PasswordValidate, так как он даёт более
//   понятные сообщения об ошибках.
func PasswordValidateRegex(password string) (bool, error) {
	// Регулярное выражение из PHP:
	// /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{10,}$/
	// Разбор:
	//   ^ - начало строки
	//   (?=.*[a-z]) - положительный lookahead: должна быть хотя бы одна строчная буква
	//   (?=.*[A-Z]) - положительный lookahead: должна быть хотя бы одна заглавная буква
	//   (?=.*\d) - положительный lookahead: должна быть хотя бы одна цифра
	//   [a-zA-Z\d]{10,} - от 10 и более символов, только буквы и цифры
	//   $ - конец строки
	pattern := `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d]{10,}$`

	// Компилируем регулярное выражение
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	// Проверяем соответствие пароля регулярному выражению
	return regex.MatchString(password), nil
}

