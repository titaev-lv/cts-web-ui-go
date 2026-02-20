// Package services содержит бизнес-логику приложения.
// Этот файл содержит сервис аутентификации (Login, Logout, Remember Me).
package services

import (
	"crypto/rand"
	"ctweb/internal/config"
	"ctweb/internal/errors"
	"ctweb/internal/logger"
	"ctweb/internal/models"
	"ctweb/internal/repositories"
	"ctweb/internal/utils"
	"encoding/hex"
	"fmt"
	"time"
)

// AuthService - сервис для аутентификации пользователей.
// Содержит бизнес-логику для входа, выхода и "Remember Me".
type AuthService struct {
	userRepo *repositories.UserRepository
	config   *config.Config
}

// NewAuthService создаёт новый экземпляр AuthService.
//
// Возвращает:
//   - *AuthService: новый сервис аутентификации
func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: repositories.NewUserRepository(),
		config:   config.Get(),
	}
}

// LoginResult представляет результат операции входа.
//
// Поля:
//   - User: данные пользователя (если вход успешен)
//   - Token: токен для "Remember Me" (если remember = true)
//   - Error: ошибка (если вход не удался)
type LoginResult struct {
	User    *models.User       `json:"user,omitempty"`
	Token   string             `json:"token,omitempty"`
	Timings map[string]float64 `json:"-"`
	Error   error              `json:"-"`
}

// Login выполняет аутентификацию пользователя.
//
// Что делает:
//  1. Находит пользователя по логину
//  2. Проверяет, что пользователь активен
//  3. Проверяет пароль через bcrypt
//  4. Проверяет, что у пользователя есть активные группы
//  5. Загружает группы пользователя
//  6. Обновляет временную метку активности
//  7. Если remember = true, генерирует токен и сохраняет его
//
// Параметры:
//   - login: логин пользователя
//   - password: пароль пользователя (в открытом виде)
//   - remember: нужно ли создать токен "Remember Me"
//
// Возвращает:
//   - *LoginResult: результат операции входа
//
// Пример использования:
//
//	authService := services.NewAuthService()
//	result := authService.Login("admin", "password123", true)
//	if result.Error != nil {
//	    // Обработка ошибки
//	}
func (s *AuthService) Login(login, password string, remember bool) *LoginResult {
	loginStart := time.Now()
	timings := map[string]float64{}
	addTiming := func(key string, startedAt time.Time) {
		timings[key] = float64(time.Since(startedAt).Microseconds()) / 1000.0
	}
	finalizeTimings := func() map[string]float64 {
		timings["auth_service_ms"] = float64(time.Since(loginStart).Microseconds()) / 1000.0
		timings["db_latency_ms"] = timings["db_find_user_ms"] + timings["db_find_groups_ms"] + timings["db_update_timestamp_ms"] + timings["db_update_token_ms"]
		return timings
	}

	// ============================================
	// ШАГ 1: Поиск пользователя по логину
	// ============================================
	findUserStart := time.Now()
	user, err := s.userRepo.FindByLogin(login)
	addTiming("db_find_user_ms", findUserStart)
	if err != nil {
		logger.Warn().
			Str("login", login).
			Err(err).
			Msg("Login attempt: user not found")
		return &LoginResult{
			Timings: finalizeTimings(),
			Error:   errors.UnauthorizedError("Bad Login or Password"),
		}
	}

	// ============================================
	// ШАГ 2: Проверка активности пользователя
	// ============================================
	if !user.IsActive() {
		// Пользователь заблокирован
		// Логируем попытку входа заблокированного пользователя
		logger.Warn().
			Str("login", login).
			Int("user_id", user.ID).
			Str("event", "failed_login").
			Str("reason", "user_blocked").
			Msg("Login attempt failed: user is blocked")
		return &LoginResult{
			Timings: finalizeTimings(),
			Error:   errors.UnauthorizedError("User is blocked"),
		}
	}

	// ============================================
	// ШАГ 3: Проверка пароля
	// ============================================
	// Сравниваем введённый пароль с хешем из БД
	// utils.PasswordVerify использует bcrypt.CompareHashAndPassword,
	// который автоматически проверяет пароль и защищает от timing attacks
	passwordVerifyStart := time.Now()
	isValid, err := utils.PasswordVerify(password, user.Password)
	addTiming("password_verify_ms", passwordVerifyStart)
	if err != nil || !isValid {
		// ВАЖНО: НЕ логируем пароль в открытом виде!
		// Логируем только метаданные для безопасности:
		// - Логин (для обнаружения брутфорс атак)
		// - User ID (если пользователь найден)
		// - IP адрес (добавится в контроллере)
		// - Время попытки
		// Это помогает обнаружить подозрительную активность без компрометации паролей
		logger.Warn().
			Str("login", login).
			Int("user_id", user.ID).
			Str("event", "failed_login").
			Str("reason", "incorrect_password").
			Msg("Login attempt failed: incorrect password")
		return &LoginResult{
			Timings: finalizeTimings(),
			Error:   errors.UnauthorizedError("Bad Login or Password"),
		}
	}

	// ============================================
	// ШАГ 4: Загрузка групп пользователя
	// ============================================
	// Загружаем только активные группы
	findGroupsStart := time.Now()
	groups, err := s.userRepo.FindGroupsByUserID(user.ID)
	addTiming("db_find_groups_ms", findGroupsStart)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", user.ID).
			Msg("Failed to load user groups")
		return &LoginResult{
			Timings: finalizeTimings(),
			Error:   errors.InternalError("Failed to load user groups", err),
		}
	}

	// Проверяем, что у пользователя есть хотя бы одна активная группа
	if len(groups) == 0 {
		logger.Warn().
			Int("user_id", user.ID).
			Str("login", login).
			Msg("Login attempt: user has no active groups")
		return &LoginResult{
			Timings: finalizeTimings(),
			Error:   errors.UnauthorizedError("User's groups is blocked or group not set"),
		}
	}

	// Устанавливаем группы пользователю
	user.SetGroups(groups)

	// ============================================
	// ШАГ 5: Обновление временной метки активности
	// ============================================
	// Обновляем TIMESTAMP_X в БД (время последнего входа)
	updateTimestampStart := time.Now()
	err = s.userRepo.UpdateTimestamp(user.ID)
	addTiming("db_update_timestamp_ms", updateTimestampStart)
	if err != nil {
		// Логируем ошибку, но не прерываем процесс входа
		logger.Warn().
			Err(err).
			Int("user_id", user.ID).
			Msg("Failed to update user timestamp")
	}

	// ============================================
	// ШАГ 6: Генерация токена для "Remember Me" (если нужно)
	// ============================================
	var token string
	if remember {
		// Генерируем случайный токен
		token, err = s.generateToken()
		if err != nil {
			logger.Error().
				Err(err).
				Int("user_id", user.ID).
				Msg("Failed to generate remember me token")
			return &LoginResult{
				Error: errors.InternalError("Failed to generate token", err),
			}
		}

		// Сохраняем токен в БД
		updateTokenStart := time.Now()
		err = s.userRepo.UpdateToken(user.ID, token)
		addTiming("db_update_token_ms", updateTokenStart)
		if err != nil {
			logger.Error().
				Err(err).
				Int("user_id", user.ID).
				Msg("Failed to save remember me token")
			return &LoginResult{
				Timings: finalizeTimings(),
				Error:   errors.InternalError("Failed to save token", err),
			}
		}
	}

	// ============================================
	// ШАГ 7: Логирование успешного входа
	// ============================================
	logger.Info().
		Int("user_id", user.ID).
		Str("login", login).
		Str("name", user.GetFullName()).
		Bool("remember_me", remember).
		Msg("User logged in successfully")

	// Возвращаем успешный результат
	return &LoginResult{
		User:    user,
		Token:   token,
		Timings: finalizeTimings(),
		Error:   nil,
	}
}

// Logout выполняет выход пользователя из системы.
//
// Что делает:
//  1. Удаляет токен "Remember Me" из БД (очищает поле TOKEN)
//  2. Логирует выход
//
// Параметры:
//   - userID: ID пользователя
//
// Возвращает:
//   - error: ошибка, если произошла ошибка БД
//
// Пример использования:
//
//	authService := services.NewAuthService()
//	err := authService.Logout(userID)
//	if err != nil {
//	    // Обработка ошибки
//	}
func (s *AuthService) Logout(userID int) error {
	// Удаляем токен из БД (устанавливаем пустую строку)
	err := s.userRepo.UpdateToken(userID, "")
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", userID).
			Msg("Failed to clear remember me token on logout")
		return errors.InternalError("Failed to logout", err)
	}

	logger.Info().
		Int("user_id", userID).
		Msg("User logged out")

	return nil
}

// AuthenticateByToken аутентифицирует пользователя по токену "Remember Me".
//
// Используется при восстановлении сессии из cookie.
//
// Что делает:
//  1. Находит пользователя по логину и токену
//  2. Проверяет, что пользователь активен
//  3. Загружает группы пользователя
//  4. Обновляет временную метку активности
//
// Параметры:
//   - login: логин из cookie
//   - token: токен из cookie
//
// Возвращает:
//   - *models.User: пользователь (если токен валиден)
//   - error: ошибка, если токен неверен или пользователь не найден
//
// Пример использования:
//
//	authService := services.NewAuthService()
//	user, err := authService.AuthenticateByToken(login, token)
//	if err != nil {
//	    // Токен неверен
//	}
func (s *AuthService) AuthenticateByToken(login, token string) (*models.User, error) {
	// ============================================
	// ШАГ 1: Поиск пользователя по логину и токену
	// ============================================
	user, err := s.userRepo.FindByLoginAndToken(login, token)
	if err != nil {
		logger.Warn().
			Str("login", login).
			Str("token", token[:min(8, len(token))]+"..."). // Логируем только первые 8 символов
			Msg("Token authentication failed: user not found")
		return nil, errors.UnauthorizedError("Token is incorrect")
	}

	// ============================================
	// ШАГ 2: Проверка активности пользователя
	// ============================================
	if !user.IsActive() {
		logger.Warn().
			Int("user_id", user.ID).
			Str("login", login).
			Msg("Token authentication failed: user is blocked")
		return nil, errors.UnauthorizedError("User is blocked")
	}

	// ============================================
	// ШАГ 3: Загрузка групп пользователя
	// ============================================
	groups, err := s.userRepo.FindGroupsByUserID(user.ID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", user.ID).
			Msg("Failed to load user groups")
		return nil, errors.InternalError("Failed to load user groups", err)
	}

	if len(groups) == 0 {
		logger.Warn().
			Int("user_id", user.ID).
			Str("login", login).
			Msg("Token authentication failed: user has no active groups")
		return nil, errors.UnauthorizedError("User's groups is blocked or group not set")
	}

	user.SetGroups(groups)

	// ============================================
	// ШАГ 4: Обновление временной метки активности
	// ============================================
	err = s.userRepo.UpdateTimestamp(user.ID)
	if err != nil {
		logger.Warn().
			Err(err).
			Int("user_id", user.ID).
			Msg("Failed to update user timestamp")
	}

	logger.Info().
		Int("user_id", user.ID).
		Str("login", login).
		Msg("User authenticated by token")

	return user, nil
}

// generateToken генерирует случайный токен для "Remember Me".
//
// Токен генерируется как случайная hex-строка длиной 64 символа (32 байта).
//
// Возвращает:
//   - string: сгенерированный токен
//   - error: ошибка, если не удалось сгенерировать токен
func (s *AuthService) generateToken() (string, error) {
	// Генерируем 32 случайных байта
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Преобразуем в hex-строку (64 символа)
	token := hex.EncodeToString(bytes)
	return token, nil
}

// min возвращает минимальное из двух чисел.
// Вспомогательная функция для безопасного логирования токена.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
