package session

import (
	"crypto/rand"
	"ctweb/internal/config"
	"ctweb/internal/logger"
	"ctweb/internal/models"
	"ctweb/internal/repositories"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

// SessionManager управляет сессиями пользователей.
// Использует cookies для хранения сессий и токенов "Remember Me".
type SessionManager struct {
	store    *sessions.CookieStore // Хранилище сессий (cookies)
	userRepo *repositories.UserRepository
	config   *config.Config
}

// SessionKeys - ключи для хранения данных в сессии
const (
	SessionKeyAuth    = "ct_auth"    // Флаг авторизации (как в PHP: $_SESSION['ct_auth'])
	SessionKeyUserID  = "ct_user_uid" // ID пользователя (как в PHP: $_SESSION['ct_user']['uid'])
	SessionKeyUserName = "ct_user_name" // Имя пользователя (как в PHP: $_SESSION['ct_user']['name'])
	SessionKeyUserEmail = "ct_user_email" // Email пользователя (как в PHP: $_SESSION['ct_user']['email'])
	SessionKeyUserGroups = "ct_user_grp" // Группы пользователя (как в PHP: $_SESSION['ct_user']['grp'])
	SessionKeyUserTimezone = "ct_user_timezone" // Часовой пояс (как в PHP: $_SESSION['ct_user']['timezone'])
)

// CookieNames - имена cookies (как в PHP)
const (
	CookieNameLogin = "Login"   // Cookie с логином (как в PHP: $_COOKIE['Login'])
	CookieNameToken = "CTToken"  // Cookie с токеном (как в PHP: $_COOKIE['CTToken'])
)

var (
	// sessionManager - глобальный экземпляр SessionManager
	sessionManager *SessionManager
)

// Init инициализирует глобальный SessionManager.
//
// Что делает:
//   1. Создаёт CookieStore с секретным ключом из конфигурации
//   2. Настраивает параметры cookies (Secure, HttpOnly, SameSite)
//   3. Создаёт экземпляр SessionManager
//
// Вызывается один раз при старте приложения (в main.go).
func Init() {
	cfg := config.Get()

	// Создаём хранилище сессий на основе cookies
	// Секретный ключ берётся из конфигурации
	store := sessions.NewCookieStore([]byte(cfg.Security.SessionSecret))

	// Настраиваем параметры cookies для безопасности
	store.Options = &sessions.Options{
		Path:     "/",                                    // Cookie доступен для всего сайта
		MaxAge:   cfg.Security.SessionMaxAge,            // Время жизни сессии (по умолчанию 24 часа)
		HttpOnly: cfg.Security.SessionCookieHTTPOnly,    // Запретить доступ через JavaScript (защита от XSS)
		Secure:   cfg.Security.SessionCookieSecure,      // Использовать только HTTPS (в продакшн)
		SameSite: getSameSiteMode(cfg.Security.SessionCookieSameSite), // Политика SameSite
	}

	// Создаём SessionManager
	sessionManager = &SessionManager{
		store:    store,
		userRepo: repositories.NewUserRepository(),
		config:   cfg,
	}

	logger.Info().Msg("Session manager initialized")
}

// getSameSiteMode преобразует строку в http.SameSite режим.
func getSameSiteMode(mode string) http.SameSite {
	switch mode {
	case "Strict":
		return http.SameSiteStrictMode
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode // По умолчанию Lax
	}
}

// GetSessionManager возвращает глобальный экземпляр SessionManager.
//
// Возвращает:
//   - *SessionManager: экземпляр менеджера сессий
//
// Паникует, если SessionManager не инициализирован (Init не был вызван).
func GetSessionManager() *SessionManager {
	if sessionManager == nil {
		panic("SessionManager not initialized. Call session.Init() first.")
	}
	return sessionManager
}

// GetSession получает сессию из запроса.
//
// Параметры:
//   - r: HTTP запрос
//   - w: HTTP ответ
//
// Возвращает:
//   - *sessions.Session: сессия пользователя
//   - error: ошибка, если не удалось получить сессию
//
// Использование:
//   session, err := sm.GetSession(r, w)
//   if err != nil {
//       // обработка ошибки
//   }
func (sm *SessionManager) GetSession(r *http.Request, w http.ResponseWriter) (*sessions.Session, error) {
	// Получаем сессию по имени из конфигурации
	// Если сессии нет, создаётся новая
	session, err := sm.store.Get(r, sm.config.Security.SessionCookieName)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

// SetUser сохраняет данные пользователя в сессию.
//
// Параметры:
//   - r: HTTP запрос
//   - w: HTTP ответ
//   - user: данные пользователя
//
// Что делает:
//   1. Получает сессию
//   2. Сохраняет данные пользователя в сессию (как в PHP: $_SESSION['ct_user'])
//   3. Устанавливает флаг авторизации (как в PHP: $_SESSION['ct_auth'] = true)
//   4. Сохраняет сессию
//
// Использование:
//   err := sm.SetUser(r, w, user)
//   if err != nil {
//       // обработка ошибки
//   }
func (sm *SessionManager) SetUser(r *http.Request, w http.ResponseWriter, user *models.User) error {
	session, err := sm.GetSession(r, w)
	if err != nil {
		return err
	}

	// Сохраняем данные пользователя в сессию (как в PHP)
	session.Values[SessionKeyAuth] = true
	session.Values[SessionKeyUserID] = user.ID
	session.Values[SessionKeyUserName] = user.GetFullName()
	session.Values[SessionKeyUserEmail] = user.Email
	session.Values[SessionKeyUserTimezone] = user.Timezone

	// Сохраняем группы пользователя
	// user.Groups уже является []int, поэтому просто сохраняем его
	session.Values[SessionKeyUserGroups] = user.Groups

	// Сохраняем сессию
	if err := session.Save(r, w); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// GetUser получает данные пользователя из сессии.
//
// Параметры:
//   - r: HTTP запрос
//
// Возвращает:
//   - *models.User: данные пользователя (если авторизован)
//   - bool: true, если пользователь авторизован
//   - error: ошибка, если произошла проблема
//
// Использование:
//   user, isAuth, err := sm.GetUser(r)
//   if err != nil {
//       // обработка ошибки
//   }
//   if isAuth {
//       // пользователь авторизован
//   }
func (sm *SessionManager) GetUser(r *http.Request) (*models.User, bool, error) {
	session, err := sm.store.Get(r, sm.config.Security.SessionCookieName)
	if err != nil {
		return nil, false, err
	}

	// Проверяем флаг авторизации (как в PHP: isset($_SESSION['ct_auth']))
	auth, ok := session.Values[SessionKeyAuth].(bool)
	if !ok || !auth {
		return nil, false, nil
	}

	// Получаем ID пользователя
	userID, ok := session.Values[SessionKeyUserID].(int)
	if !ok || userID == 0 {
		return nil, false, nil
	}

	// Загружаем пользователя из БД
	user, err := sm.userRepo.FindByID(userID)
	if err != nil {
		return nil, false, err
	}

	// Загружаем группы пользователя
	groups, err := sm.userRepo.FindGroupsByUserID(userID)
	if err != nil {
		return nil, false, err
	}
	user.Groups = groups

	return user, true, nil
}

// ClearUser очищает сессию пользователя (выход из системы).
//
// Параметры:
//   - r: HTTP запрос
//   - w: HTTP ответ
//
// Что делает:
//   1. Получает сессию
//   2. Очищает все значения сессии
//   3. Удаляет cookies "Remember Me" (Login и CTToken)
//   4. Сохраняет сессию
//
// Использование:
//   err := sm.ClearUser(r, w)
//   if err != nil {
//       // обработка ошибки
//   }
func (sm *SessionManager) ClearUser(r *http.Request, w http.ResponseWriter) error {
	session, err := sm.GetSession(r, w)
	if err != nil {
		return err
	}

	// Очищаем все значения сессии
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1 // Удалить cookie

	// Удаляем cookies "Remember Me" (как в PHP)
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameLogin,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   sm.config.Security.SessionCookieSecure,
		SameSite: getSameSiteMode(sm.config.Security.SessionCookieSameSite),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameToken,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   sm.config.Security.SessionCookieSecure,
		SameSite: getSameSiteMode(sm.config.Security.SessionCookieSameSite),
	})

	// Сохраняем сессию (это удалит cookie)
	if err := session.Save(r, w); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// SetRememberMeCookies устанавливает cookies "Remember Me".
//
// Параметры:
//   - w: HTTP ответ
//   - login: логин пользователя
//   - token: токен из БД
//
// Что делает:
//   1. Устанавливает cookie "Login" с логином
//   2. Устанавливает cookie "CTToken" с токеном
//   3. Срок действия: RememberMeDays дней (по умолчанию 7 дней)
//
// Использование:
//   sm.SetRememberMeCookies(w, "admin", "abc123...")
func (sm *SessionManager) SetRememberMeCookies(w http.ResponseWriter, login, token string) {
	// Вычисляем время истечения (как в PHP: +7 дней)
	expires := time.Now().Add(time.Duration(sm.config.Security.RememberMeDays) * 24 * time.Hour)

	// Устанавливаем cookie "Login" (как в PHP: setcookie("Login", $login, ...))
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameLogin,
		Value:    login,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true, // Защита от XSS
		Secure:   sm.config.Security.SessionCookieSecure,
		SameSite: getSameSiteMode(sm.config.Security.SessionCookieSameSite),
	})

	// Устанавливаем cookie "CTToken" (как в PHP: setcookie("CTToken", $h, ...))
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameToken,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true, // Защита от XSS
		Secure:   sm.config.Security.SessionCookieSecure,
		SameSite: getSameSiteMode(sm.config.Security.SessionCookieSameSite),
	})
}

// RestoreUserFromCookies восстанавливает пользователя из cookies "Remember Me".
//
// Параметры:
//   - r: HTTP запрос
//   - w: HTTP ответ
//
// Что делает:
//   1. Читает cookies "Login" и "CTToken"
//   2. Проверяет токен в БД (как в PHP: WHERE LOGIN = ? AND TOKEN = ?)
//   3. Если токен валиден, восстанавливает сессию
//   4. Загружает группы пользователя
//
// Возвращает:
//   - *models.User: данные пользователя (если токен валиден)
//   - bool: true, если пользователь восстановлен
//   - error: ошибка, если произошла проблема
//
// Использование:
//   user, restored, err := sm.RestoreUserFromCookies(r, w)
//   if err != nil {
//       // обработка ошибки
//   }
//   if restored {
//       // пользователь восстановлен из cookies
//   }
func (sm *SessionManager) RestoreUserFromCookies(r *http.Request, w http.ResponseWriter) (*models.User, bool, error) {
	// Читаем cookies (как в PHP: isset($_COOKIE['Login']) && isset($_COOKIE['CTToken']))
	loginCookie, err := r.Cookie(CookieNameLogin)
	if err != nil {
		return nil, false, nil // Cookie не найдена - это нормально
	}

	tokenCookie, err := r.Cookie(CookieNameToken)
	if err != nil {
		return nil, false, nil // Cookie не найдена - это нормально
	}

	login := loginCookie.Value
	token := tokenCookie.Value

	if login == "" || token == "" {
		return nil, false, nil
	}

	// Ищем пользователя по логину и токену (как в PHP)
	user, err := sm.userRepo.FindByLoginAndToken(login, token)
	if err != nil {
		// Токен не найден или неверный
		logger.Warn().
			Str("login", login).
			Str("event", "invalid_remember_me_token").
			Msg("Invalid Remember Me token")
		return nil, false, nil
	}

	// Проверяем, что пользователь активен
	if !user.IsActive() {
		logger.Warn().
			Str("login", login).
			Int("user_id", user.ID).
			Str("event", "user_blocked").
			Msg("User is blocked, cannot restore from Remember Me")
		return nil, false, nil
	}

	// Загружаем группы пользователя
	groups, err := sm.userRepo.FindGroupsByUserID(user.ID)
	if err != nil {
		return nil, false, err
	}
	user.Groups = groups

	// Проверяем, что у пользователя есть активные группы
	if len(groups) == 0 {
		logger.Warn().
			Int("user_id", user.ID).
			Str("event", "no_active_groups").
			Msg("User has no active groups, cannot restore from Remember Me")
		return nil, false, nil
	}

	// Восстанавливаем сессию
	if err := sm.SetUser(r, w, user); err != nil {
		return nil, false, err
	}

	logger.Info().
		Str("login", login).
		Int("user_id", user.ID).
		Str("event", "user_restored_from_cookies").
		Msg("User restored from Remember Me cookies")

	return user, true, nil
}

// GenerateRememberMeToken генерирует случайный токен для "Remember Me".
//
// Возвращает:
//   - string: токен в hex формате
//   - error: ошибка, если не удалось сгенерировать токен
//
// Использование:
//   token, err := sm.GenerateRememberMeToken()
//   if err != nil {
//       // обработка ошибки
//   }
//
// Примечание:
//   Генерирует токен длиной 32 байта (256 бит) в hex формате.
//   Это безопаснее, чем в PHP (где используется mt_rand(2, 100) байт).
func (sm *SessionManager) GenerateRememberMeToken() (string, error) {
	// Генерируем случайные байты (как в PHP: openssl_random_pseudo_bytes)
	// В PHP используется длина от 2 до 100 байт (mt_rand(2, 100))
	// Для безопасности используем фиксированную длину 32 байта (256 бит)
	tokenBytes := make([]byte, 32)

	// Используем crypto/rand для генерации безопасных случайных байт
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Преобразуем в hex строку (как в PHP: bin2hex)
	token := hex.EncodeToString(tokenBytes)

	return token, nil
}

