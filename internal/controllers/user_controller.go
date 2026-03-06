// Package controllers содержит HTTP handlers (контроллеры) для обработки запросов.
// Контроллеры получают запросы, вызывают сервисы и возвращают ответы.
package controllers

import (
	"crypto/sha256"
	"ctweb/internal/config"
	"ctweb/internal/errors" // Централизованная обработка ошибок
	"ctweb/internal/hsm"
	"ctweb/internal/logger"       // Система логирования
	"ctweb/internal/middleware"   // Middleware для получения пользователя
	"ctweb/internal/models"       // Модели данных
	"ctweb/internal/repositories" // Репозитории для работы с БД
	"ctweb/internal/services"
	"ctweb/internal/session" // Управление сессиями
	"ctweb/internal/utils"   // Утилиты (DataTables парсинг, валидация)
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// UserController - контроллер для работы с пользователями.
// Содержит методы для обработки HTTP запросов, связанных с пользователями.
type UserController struct{}

var (
	hsmClientOnce sync.Once
	hsmClientInst *hsm.Client
	hsmClientErr  error
)

// NewUserController создаёт новый экземпляр UserController.
//
// Возвращает:
//   - *UserController: новый контроллер
func NewUserController() *UserController {
	return &UserController{}
}

// Home обрабатывает запрос на главную страницу.
//
// Параметры:
//   - c: контекст Gin (содержит запрос и ответ)
//
// Что делает:
//
//	Рендерит HTML шаблон главной страницы с базовым layout.
//	Передаёт информацию о пользователе (если авторизован) в шаблон.
//
// Шаблоны:
//   - index.html - контент главной страницы
//
// Home отображает главную страницу системы.
// Получает текущего пользователя из контекста (установлен в AuthMiddleware)
func (u *UserController) Home(c *gin.Context) {
	// Получаем текущего пользователя из контекста
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		// Не должно произойти, т.к. middleware проверяет аутентификацию
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userRepo := repositories.NewUserRepository()
	twoFAEnabled, err := userRepo.Is2FAEnabled(user.ID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", user.ID).
			Msg("Failed to load user 2FA status for home page")
		twoFAEnabled = false
	}

	// Рендерим template с данными пользователя
	c.HTML(http.StatusOK, "home", gin.H{
		"Title":        "Home",
		"User":         user,
		"TwoFAEnabled": twoFAEnabled,
	})
}

// UserProfile отображает страницу профиля пользователя.
// Доступна любому авторизованному пользователю.
func (u *UserController) UserProfile(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userRepo := repositories.NewUserRepository()
	twoFAEnabled, err := userRepo.Is2FAEnabled(user.ID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", user.ID).
			Msg("Failed to load user 2FA status")
		twoFAEnabled = false
	}

	c.HTML(http.StatusOK, "user_profile/index.html", gin.H{
		"Title":        "User Profile",
		"User":         user,
		"Timezones":    utils.GetTimezoneGroups(),
		"TwoFAEnabled": twoFAEnabled,
	})
}

// AjaxUpdateProfileTimezone обновляет timezone текущего пользователя.
func (u *UserController) AjaxUpdateProfileTimezone(c *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusOK, gin.H{"error": "Unauthorized", "success": false})
		return
	}

	timezone := strings.TrimSpace(c.PostForm("timezone"))
	if timezone == "" {
		c.JSON(http.StatusOK, gin.H{"error": "Field \"TimeZone\" is empty", "success": false})
		return
	}

	if _, err := time.LoadLocation(timezone); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "timezone failed", "success": false})
		return
	}

	userRepo := repositories.NewUserRepository()
	existingUser, err := userRepo.FindByID(currentUser.ID)
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to find user for timezone update")
		c.JSON(http.StatusOK, gin.H{"error": "User not found", "success": false})
		return
	}

	existingUser.Timezone = timezone
	if err := userRepo.Update(existingUser, nil, false, currentUser.ID); err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to update user timezone")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to save timezone", "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
		"data": gin.H{
			"timezone": timezone,
		},
	})
}

// AjaxUpdateProfileInfo обновляет name/last_name/email текущего пользователя.
func (u *UserController) AjaxUpdateProfileInfo(c *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusOK, gin.H{"error": "Unauthorized", "success": false})
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	lastName := strings.TrimSpace(c.PostForm("last_name"))
	email := strings.TrimSpace(c.PostForm("email"))

	if name == "" {
		c.JSON(http.StatusOK, gin.H{"error": "Field \"Name\" is empty", "success": false})
		return
	}
	if lastName == "" {
		c.JSON(http.StatusOK, gin.H{"error": "Field \"Last Name\" is empty", "success": false})
		return
	}
	if email == "" {
		c.JSON(http.StatusOK, gin.H{"error": "Field \"Email\" is empty", "success": false})
		return
	}

	if err := utils.ValidateEmail(email); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error(), "success": false})
		return
	}

	userRepo := repositories.NewUserRepository()
	existsEmail, err := userRepo.ExistsByEmailExcludingID(email, currentUser.ID)
	if err != nil {
		logger.Error().Err(err).Str("email", email).Int("user_id", currentUser.ID).Msg("Failed to check email uniqueness")
		c.JSON(http.StatusOK, gin.H{"error": "Database error", "success": false})
		return
	}
	if existsEmail {
		c.JSON(http.StatusOK, gin.H{"error": "Email already exist", "success": false})
		return
	}

	existingUser, err := userRepo.FindByID(currentUser.ID)
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to find user for profile update")
		c.JSON(http.StatusOK, gin.H{"error": "User not found", "success": false})
		return
	}

	existingUser.Name = name
	existingUser.LastName = lastName
	existingUser.Email = email

	if err := userRepo.Update(existingUser, nil, false, currentUser.ID); err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to update profile info")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to save profile", "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
		"data": gin.H{
			"name":      name,
			"last_name": lastName,
			"email":     email,
		},
	})
}

// AjaxUpdateProfilePassword обновляет пароль текущего пользователя.
func (u *UserController) AjaxUpdateProfilePassword(c *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusOK, gin.H{"error": "Unauthorized", "success": false})
		return
	}

	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("new_password_confirm")

	if currentPassword == "" {
		c.JSON(http.StatusOK, gin.H{"error": "Field \"Current Password\" is empty", "success": false})
		return
	}
	if newPassword == "" {
		c.JSON(http.StatusOK, gin.H{"error": "Field \"New Password\" is empty", "success": false})
		return
	}
	if confirmPassword == "" {
		c.JSON(http.StatusOK, gin.H{"error": "Field \"Confirm New Password\" is empty", "success": false})
		return
	}

	if err := utils.PasswordValidateWithConfirm(newPassword, confirmPassword); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error(), "success": false})
		return
	}

	userRepo := repositories.NewUserRepository()
	existingUser, err := userRepo.FindByID(currentUser.ID)
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to find user for password update")
		c.JSON(http.StatusOK, gin.H{"error": "User not found", "success": false})
		return
	}

	passwordOk, err := utils.PasswordVerify(currentPassword, existingUser.Password)
	if err != nil || !passwordOk {
		c.JSON(http.StatusOK, gin.H{"error": "Current password is incorrect", "success": false})
		return
	}

	hashedPassword, err := utils.PasswordHash(newPassword)
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to hash new password")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to hash password", "success": false})
		return
	}

	existingUser.Password = hashedPassword
	if err := userRepo.Update(existingUser, nil, true, currentUser.ID); err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to update password")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to save password", "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"error": false, "success": true})
}

// AjaxUpdateProfile2FA включает/выключает 2FA для текущего пользователя.
func (u *UserController) AjaxUpdateProfile2FA(c *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusOK, gin.H{"error": "Unauthorized", "success": false})
		return
	}

	enabledRaw := strings.TrimSpace(strings.ToLower(c.PostForm("enabled")))
	enabled := enabledRaw == "1" || enabledRaw == "true" || enabledRaw == "on" || enabledRaw == "yes"
	if enabled {
		c.JSON(http.StatusOK, gin.H{"error": "Use 2FA setup flow to enable", "success": false})
		return
	}

	userRepo := repositories.NewUserRepository()
	if err := userRepo.DisableAndClear2FA(currentUser.ID); err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Bool("enabled", enabled).Msg("Failed to update 2FA status")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to save 2FA settings", "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
		"data": gin.H{
			"enabled": false,
		},
	})
}

// AjaxStartProfile2FASetup starts 2FA enrollment: generates TOTP secret, encrypts via HSM, saves pending state.
func (u *UserController) AjaxStartProfile2FASetup(c *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusOK, gin.H{"error": "Unauthorized", "success": false})
		return
	}

	hsmClient, hsmContext, err := getHSM2FAClient()
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("2FA setup failed: HSM unavailable")
		c.JSON(http.StatusOK, gin.H{"error": "2FA service unavailable", "success": false})
		return
	}

	secret, err := utils.GenerateTOTPSecret(20)
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to generate TOTP secret")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to start 2FA", "success": false})
		return
	}

	keyID, ciphertext, err := hsmClient.Encrypt(c.Request.Context(), hsmContext, []byte(secret))
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to encrypt TOTP secret via HSM")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to start 2FA", "success": false})
		return
	}

	keyVersion := parseHSMKeyVersion(keyID)
	if keyVersion <= 0 {
		keyVersion = 1
	}

	userRepo := repositories.NewUserRepository()
	if err := userRepo.Upsert2FASecret(currentUser.ID, ciphertext, keyVersion); err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to persist 2FA setup data")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to start 2FA", "success": false})
		return
	}

	issuer := "CT-System"
	account := strings.TrimSpace(currentUser.Login)
	if account == "" {
		account = strings.TrimSpace(currentUser.Email)
	}
	otpAuthURI := utils.BuildOTPAuthURI(secret, issuer, account)
	qrCodeDataURI, err := utils.GenerateQRCodeDataURI(otpAuthURI, 256)
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to generate 2FA QR code")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to start 2FA", "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
		"data": gin.H{
			"secret":  secret,
			"qr_code": qrCodeDataURI,
		},
	})
}

// AjaxVerifyProfile2FASetup verifies TOTP code and enables 2FA.
func (u *UserController) AjaxVerifyProfile2FASetup(c *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusOK, gin.H{"error": "Unauthorized", "success": false})
		return
	}

	code := strings.TrimSpace(c.PostForm("code"))
	if code == "" {
		c.JSON(http.StatusOK, gin.H{"error": "Field \"Verification Code\" is empty", "success": false})
		return
	}

	hsmClient, hsmContext, err := getHSM2FAClient()
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("2FA verify failed: HSM unavailable")
		c.JSON(http.StatusOK, gin.H{"error": "2FA service unavailable", "success": false})
		return
	}

	userRepo := repositories.NewUserRepository()
	rec, err := userRepo.Get2FARecord(currentUser.ID)
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to load 2FA record")
		c.JSON(http.StatusOK, gin.H{"error": "2FA setup not found", "success": false})
		return
	}
	if rec == nil || strings.TrimSpace(rec.SecretEnc) == "" {
		c.JSON(http.StatusOK, gin.H{"error": "2FA setup not started", "success": false})
		return
	}

	keyVersion := rec.EncKeyVersion
	if keyVersion <= 0 {
		keyVersion = 1
	}
	keyID := "kek-2fa-v" + strconv.Itoa(keyVersion)

	plaintext, err := hsmClient.Decrypt(c.Request.Context(), hsmContext, keyID, rec.SecretEnc)
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Int("key_version", keyVersion).Msg("Failed to decrypt 2FA secret")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to verify 2FA code", "success": false})
		return
	}

	secret := strings.TrimSpace(string(plaintext))
	if !utils.VerifyTOTPCode(secret, code, time.Now(), 1) {
		c.JSON(http.StatusOK, gin.H{"error": "Invalid verification code", "success": false})
		return
	}

	recoveryCodes, err := utils.GenerateRecoveryCodes(10)
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to generate recovery codes")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to enable 2FA", "success": false})
		return
	}

	recoveryCodesJSON, err := json.Marshal(recoveryCodes)
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to marshal recovery codes")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to enable 2FA", "success": false})
		return
	}

	recoveryKeyID, recoveryCiphertext, err := hsmClient.Encrypt(c.Request.Context(), hsmContext, recoveryCodesJSON)
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to encrypt recovery codes")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to enable 2FA", "success": false})
		return
	}

	recoveryKeyVersion := parseHSMKeyVersion(recoveryKeyID)
	if recoveryKeyVersion <= 0 {
		recoveryKeyVersion = keyVersion
	}

	recoveryHashes := utils.HashRecoveryCodes(recoveryCodes)
	recoveryHashesJSON, err := json.Marshal(recoveryHashes)
	if err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to marshal recovery code hashes")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to enable 2FA", "success": false})
		return
	}

	if err := userRepo.Enable2FAWithRecovery(currentUser.ID, recoveryCiphertext, string(recoveryHashesJSON), recoveryKeyVersion); err != nil {
		logger.Error().Err(err).Int("user_id", currentUser.ID).Msg("Failed to enable 2FA")
		c.JSON(http.StatusOK, gin.H{"error": "Failed to enable 2FA", "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
		"data": gin.H{
			"enabled":        true,
			"recovery_codes": recoveryCodes,
		},
	})
}

func getHSM2FAClient() (*hsm.Client, string, error) {
	cfg := config.Get()
	if !cfg.HSM.Enabled {
		return nil, "", fmt.Errorf("hsm is disabled")
	}

	hsmClientOnce.Do(func() {
		hsmClientInst, hsmClientErr = hsm.NewClient(hsm.ClientConfig{
			BaseURL:        cfg.HSM.URL,
			CertPath:       cfg.HSM.TLS.CertPath,
			KeyPath:        cfg.HSM.TLS.KeyPath,
			CAPath:         cfg.HSM.TLS.CAPath,
			RequestTimeout: cfg.HSM.Timeout,
			RetryConfig: hsm.RetryConfig{
				MaxAttempts: cfg.HSM.Retry.MaxAttempts,
				InitialWait: cfg.HSM.Retry.InitialDelay,
				MaxWait:     cfg.HSM.Retry.MaxDelay,
				Multiplier:  cfg.HSM.Retry.Multiplier,
			},
		})
	})

	if hsmClientErr != nil {
		return nil, "", hsmClientErr
	}
	return hsmClientInst, cfg.HSM.Context, nil
}

func parseHSMKeyVersion(keyID string) int {
	re := regexp.MustCompile(`v(\d+)$`)
	match := re.FindStringSubmatch(strings.TrimSpace(keyID))
	if len(match) != 2 {
		return 0
	}
	v, err := strconv.Atoi(match[1])
	if err != nil {
		return 0
	}
	return v
}

// ShowLoginPage отображает страницу входа.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//
//	Рендерит HTML шаблон login.html.
//
// Шаблон должен находиться в web/templates/login.html
func (u *UserController) ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

// ShowLogin2FAPage рендерит отдельную страницу второго шага входа.
func (u *UserController) ShowLogin2FAPage(c *gin.Context) {
	sm := session.GetSessionManager()
	_, _, _, hasPending2FA, err := sm.GetPending2FA(c.Request)
	if err != nil || !hasPending2FA {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "login_2fa.html", nil)
}

// Login обрабатывает запрос на вход пользователя.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//  1. Парсит JSON с логином и паролем
//  2. Проверяет креды пользователя
//  3. Создает session/cookie и remember-me токен
//  4. Возвращает ответ в формате, совместимом с PHP
//
// Формат запроса:
//
//	POST /auth/login
//	{
//	  "username": "user123",
//	  "password": "password123"
//	}
//
// Формат ответа (успех):
//
//	{
//	  "error": false,
//	  "success": true,
//	  "data": {
//	    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
//	  }
//	}
//
// Формат ответа (ошибка):
//
//	{
//	  "error": "Invalid credentials",
//	  "success": false
//	}
func (u *UserController) Login(c *gin.Context) {
	// ============================================
	// ШАГ 1: Парсинг form data (как в PHP)
	// ============================================
	// PHP форма отправляет данные как application/x-www-form-urlencoded
	// Поля: username, pwd, rememberme
	username := c.PostForm("username")           // Логин пользователя
	password := c.PostForm("pwd")                // Пароль пользователя (в PHP это 'pwd', не 'password')
	remember := c.PostForm("rememberme") == "on" // Remember Me (checkbox)
	otpCode := strings.TrimSpace(c.PostForm("otp_code"))
	recoveryCode := strings.TrimSpace(c.PostForm("recovery_code"))
	oneTimeCode := strings.TrimSpace(c.PostForm("twofa_code"))
	if oneTimeCode != "" && otpCode == "" && recoveryCode == "" {
		// Single-field 2FA form: one value can be OTP or recovery code.
		otpCode = oneTimeCode
		recoveryCode = oneTimeCode
	}

	sm := session.GetSessionManager()
	pendingUserID, pendingRemember, pendingToken, hasPending2FA, pendingErr := sm.GetPending2FA(c.Request)
	if pendingErr != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Failed to load 2FA session",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 1.5: Завершение pending 2FA авторизации
	// ============================================
	if hasPending2FA {
		if otpCode == "" && recoveryCode == "" {
			c.JSON(http.StatusOK, gin.H{
				"error":        "Enter one-time code or recovery code",
				"success":      false,
				"requires_2fa": true,
			})
			return
		}

		ok, err := verifyLoginSecondFactor(c, pendingUserID, otpCode, recoveryCode)
		if err != nil {
			logger.Error().Err(err).Int("user_id", pendingUserID).Msg("2FA verification failed")
			c.JSON(http.StatusOK, gin.H{
				"error":        "Failed to verify second factor",
				"success":      false,
				"requires_2fa": true,
			})
			return
		}
		if !ok {
			c.JSON(http.StatusOK, gin.H{
				"error":        "Invalid one-time code or recovery code",
				"success":      false,
				"requires_2fa": true,
			})
			return
		}

		userRepo := repositories.NewUserRepository()
		user, err := userRepo.FindByID(pendingUserID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "User not found", "success": false})
			return
		}
		groups, err := userRepo.FindGroupsByUserID(pendingUserID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "Failed to load user groups", "success": false})
			return
		}
		user.SetGroups(groups)

		if err := sm.SetUser(c.Request, c.Writer, user); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "Failed to create session", "success": false})
			return
		}

		if pendingRemember && pendingToken != "" {
			sm.SetRememberMeCookies(c.Request, c.Writer, user.Login, pendingToken)
		}
		_ = sm.ClearPending2FA(c.Request, c.Writer)

		logger.Info().
			Str("login", user.Login).
			Int("user_id", user.ID).
			Str("client_ip", c.ClientIP()).
			Str("event", "login_success_2fa").
			Msg("User completed 2FA and logged in")

		c.JSON(http.StatusOK, gin.H{"error": false, "success": true})
		return
	}

	// ============================================
	// ШАГ 2: Валидация входных данных
	// ============================================

	// Проверяем, что логин и пароль не пустые
	if username == "" {
		errors.HandleError(c, errors.ValidationError("Username is required", nil))
		return
	}

	if password == "" {
		errors.HandleError(c, errors.ValidationError("Password is required", nil))
		return
	}

	// ============================================
	// ШАГ 3: Аутентификация через AuthService
	// ============================================
	// Используем AuthService для входа (он использует bcrypt и загружает группы)
	authService := services.NewAuthService()
	authServiceStart := time.Now()
	result := authService.Login(username, password, remember)
	middleware.AddLatencyPart(c, "auth_service_call_ms", time.Since(authServiceStart))
	for key, value := range result.Timings {
		middleware.AddLatencyPartMS(c, key, value)
	}

	// Проверяем результат
	if result.Error != nil {
		// Дополнительное логирование в контроллере с IP адресом и User-Agent
		// Это помогает отслеживать подозрительную активность
		// ВАЖНО: НЕ логируем пароль!
		logger.Warn().
			Str("login", username).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.GetHeader("User-Agent")).
			Str("error", result.Error.Error()).
			Str("event", "login_failed").
			Msg("Login failed")

		// Формат ответа как в PHP: {"error": "текст ошибки", "success": false}
		// Получаем текст ошибки из AppError
		errorMsg := result.Error.Error()
		c.JSON(http.StatusOK, gin.H{
			"error":   errorMsg,
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 3.5: Проверка необходимости второго фактора
	// ============================================
	userRepo := repositories.NewUserRepository()
	twoFAEnabled, err := userRepo.Is2FAEnabled(result.User.ID)
	if err != nil {
		logger.Error().Err(err).Int("user_id", result.User.ID).Msg("Failed to check 2FA status during login")
		c.JSON(http.StatusOK, gin.H{"error": "Database error", "success": false})
		return
	}

	if twoFAEnabled {
		if err := sm.SetPending2FA(c.Request, c.Writer, result.User.ID, remember, result.Token); err != nil {
			logger.Error().Err(err).Int("user_id", result.User.ID).Msg("Failed to save pending 2FA session")
			c.JSON(http.StatusOK, gin.H{"error": "Failed to start 2FA challenge", "success": false})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"error":        false,
			"success":      false,
			"requires_2fa": true,
		})
		return
	}

	// ============================================
	// ШАГ 4: Установка сессии и cookies
	// ============================================
	// Сохраняем данные пользователя в сессию (как в PHP: $_SESSION['ct_user'])
	sessionStart := time.Now()
	if err := sm.SetUser(c.Request, c.Writer, result.User); err != nil {
		middleware.AddLatencyPart(c, "session_write_ms", time.Since(sessionStart))
		logger.Error().
			Err(err).
			Int("user_id", result.User.ID).
			Msg("Failed to set session")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Failed to create session",
			"success": false,
		})
		return
	}

	// Если выбрано "Remember Me", устанавливаем cookies с токеном
	if remember && result.Token != "" {
		rememberStart := time.Now()
		sm.SetRememberMeCookies(c.Request, c.Writer, username, result.Token)
		middleware.AddLatencyPart(c, "remember_cookie_ms", time.Since(rememberStart))
		logger.Debug().
			Str("login", username).
			Msg("Remember Me cookies set")
	}

	middleware.AddLatencyPart(c, "session_write_ms", time.Since(sessionStart))

	// ============================================
	// ШАГ 5: Логирование успешного входа
	// ============================================
	// Логируем успешный вход с метаданными для аудита
	// ВАЖНО: НЕ логируем пароль!
	logger.Info().
		Str("login", username).
		Int("user_id", result.User.ID).
		Str("client_ip", c.ClientIP()).
		Str("user_agent", c.GetHeader("User-Agent")).
		Bool("remember_me", remember).
		Str("event", "login_success").
		Msg("User logged in successfully")

	// ============================================
	// ШАГ 6: Отправка успешного ответа
	// ============================================
	// Формат ответа как в PHP: {"error": false, "success": true}
	// В PHP не возвращается токен в ответе, он сохраняется в сессии/куках
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
	})
}

func verifyLoginSecondFactor(c *gin.Context, userID int, otpCode, recoveryCode string) (bool, error) {
	userRepo := repositories.NewUserRepository()
	rec, err := userRepo.Get2FARecord(userID)
	if err != nil {
		return false, err
	}
	if rec == nil || !rec.Enabled || strings.TrimSpace(rec.SecretEnc) == "" {
		return false, nil
	}

	hsmClient, hsmContext, err := getHSM2FAClient()
	if err != nil {
		return false, err
	}

	keyVersion := rec.EncKeyVersion
	if keyVersion <= 0 {
		keyVersion = 1
	}
	keyID := "kek-2fa-v" + strconv.Itoa(keyVersion)

	plaintext, err := hsmClient.Decrypt(c.Request.Context(), hsmContext, keyID, rec.SecretEnc)
	if err != nil {
		return false, err
	}

	secret := strings.TrimSpace(string(plaintext))
	if otpCode != "" && utils.VerifyTOTPCode(secret, otpCode, time.Now(), 1) {
		return true, nil
	}

	if recoveryCode == "" {
		return false, nil
	}

	normalizedCode := normalizeRecoveryCode(recoveryCode)
	if normalizedCode == "" {
		return false, nil
	}

	hash := sha256.Sum256([]byte(normalizedCode))
	hashHex := hex.EncodeToString(hash[:])

	remaining := make([]string, 0, len(rec.RecoveryHashes))
	consumed := false
	for _, h := range rec.RecoveryHashes {
		if !consumed && strings.EqualFold(strings.TrimSpace(h), hashHex) {
			consumed = true
			continue
		}
		remaining = append(remaining, h)
	}

	if !consumed {
		return false, nil
	}

	remainingJSON, err := json.Marshal(remaining)
	if err != nil {
		return false, err
	}

	if err := userRepo.UpdateRecoveryCodeHashes(userID, string(remainingJSON)); err != nil {
		return false, err
	}

	return true, nil
}

func normalizeRecoveryCode(code string) string {
	clean := strings.ToUpper(strings.TrimSpace(code))
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.ReplaceAll(clean, " ", "")
	if len(clean) != 8 {
		return ""
	}
	return clean[:4] + "-" + clean[4:]
}

// Logout обрабатывает запрос на выход из системы.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//  1. Получает пользователя из контекста (установлен в AuthMiddleware)
//  2. Вызывает AuthService.Logout() для удаления токена из БД
//  3. Вызывает SessionManager.ClearUser() для очистки сессии и cookies
//  4. Редиректит на страницу входа
//
// Использование:
//
//	GET /auth/logout - выход из системы
//
// Примечание:
//
//	После выхода пользователь будет перенаправлен на /login
func (u *UserController) Logout(c *gin.Context) {
	// ============================================
	// ШАГ 1: Получение пользователя из контекста
	// ============================================
	// Пользователь должен быть установлен в AuthMiddleware
	// Если пользователя нет, всё равно выполняем выход (на случай частичной сессии)
	user, exists := middleware.GetUserFromContext(c)

	var userID int
	if exists && user != nil {
		userID = user.ID
		logger.Debug().
			Int("user_id", userID).
			Str("login", user.Login).
			Msg("Logout requested")
	} else {
		logger.Debug().
			Msg("Logout requested but user not found in context")
	}

	// ============================================
	// ШАГ 2: Удаление токена из БД
	// ============================================
	// Если пользователь найден, удаляем токен "Remember Me" из БД
	// Это делает невозможным восстановление сессии из cookies
	if exists && user != nil {
		authService := services.NewAuthService()
		err := authService.Logout(userID)
		if err != nil {
			// Ошибка при удалении токена - логируем, но продолжаем выход
			// Сессия всё равно будет очищена
			logger.Error().
				Err(err).
				Int("user_id", userID).
				Msg("Failed to clear remember me token on logout")
		}
	}

	// ============================================
	// ШАГ 3: Очистка сессии и cookies
	// ============================================
	// Очищаем сессию и удаляем cookies "Remember Me" (Login и CTToken)
	// Это делается всегда, даже если пользователь не найден в контексте
	sm := session.GetSessionManager()
	if err := sm.ClearUser(c.Request, c.Writer); err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to clear session on logout")
		// Продолжаем выход, даже если не удалось очистить сессию
	}

	// ============================================
	// ШАГ 4: Логирование успешного выхода
	// ============================================
	if exists && user != nil {
		logger.Info().
			Int("user_id", userID).
			Str("login", user.Login).
			Str("client_ip", c.ClientIP()).
			Str("event", "logout_success").
			Msg("User logged out successfully")
	}

	// ============================================
	// ШАГ 5: Редирект на страницу входа
	// ============================================
	// Редиректим на страницу входа (как в PHP после logout)
	c.Redirect(http.StatusFound, "/login")
}

// AjaxGetUsers обрабатывает AJAX запрос от DataTables для получения списка пользователей.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//  1. Проверяет, что пользователь - администратор
//  2. Парсит параметры DataTables запроса
//  3. Вызывает репозиторий для получения данных
//  4. Форматирует ответ в формате DataTables (aaData)
//
// Формат запроса:
//
//	POST /users/ajax_get_users
//	Параметры DataTables (draw, start, length, search[value], order[0][column], columns[i][data] и т.д.)
//
// Формат ответа:
//
//	{
//	  "draw": 1,
//	  "recordsTotal": 100,
//	  "recordsFiltered": 50,
//	  "aaData": [
//	    {
//	      "chbx": "",
//	      "DT_RowId": "row_1",
//	      "id": 1,
//	      "login": "admin",
//	      "groups": "Admin",
//	      "active": "Active",
//	      "name": "Doe John",
//	      "email": "admin@example.com",
//	      "create_date": "01-12-2025 10:30:00",
//	      "modify_date": "02-12-2025 15:20:00",
//	      "timestamp": "03-12-2025 09:15:00"
//	    }
//	  ]
//	}
func (u *UserController) AjaxGetUsers(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// ============================================
	// ШАГ 2: Парсинг параметров DataTables запроса
	// ============================================
	req := utils.ParseDataTablesRequest(c)

	// ============================================
	// ШАГ 3: Получение данных из репозитория
	// ============================================
	userRepo := repositories.NewUserRepository()

	// Конвертируем запрос в формат репозитория
	userReq := utils.ConvertToUserRepositoryRequest(req)

	// Получаем данные
	repoResponse, err := userRepo.FindAllWithPagination(userReq)
	if err != nil {
		logger.Error().
			Err(err).
			Str("client_ip", c.ClientIP()).
			Msg("Failed to get users list for DataTables")

		// Возвращаем пустой ответ в случае ошибки (как в PHP)
		c.JSON(http.StatusOK, gin.H{
			"recordsTotal":    0,
			"recordsFiltered": 0,
			"aaData":          []interface{}{},
		})
		return
	}

	// ============================================
	// ШАГ 4: Форматирование ответа в формат DataTables
	// ============================================
	response := utils.ConvertUserResponseToDataTablesFormat(req.Draw, repoResponse)

	// ============================================
	// ШАГ 5: Возврат ответа
	// ============================================
	c.JSON(http.StatusOK, response)
}

// List отображает страницу управления пользователями.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//  1. Проверяет, что пользователь - администратор
//  2. Рендерит HTML шаблон страницы пользователей
//
// Формат запроса:
//
//	GET /users/
//
// Шаблон:
//   - users/index.html (если будет создан)
func (u *UserController) List(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// ============================================
	// ШАГ 2: Загрузка всех групп для выпадающих списков
	// ============================================
	groupRepo := repositories.NewGroupRepository()
	groups, err := groupRepo.FindAll()
	if err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to load groups for users page")
		// Продолжаем с пустым списком групп
		groups = []*models.Group{}
	}

	// ============================================
	// ШАГ 3: Рендеринг страницы
	// ============================================
	c.HTML(http.StatusOK, "users/index.html", gin.H{
		"Title":     "Users Management",
		"User":      user,
		"Groups":    groups,
		"Timezones": utils.GetTimezoneGroups(),
	})
}

// AjaxGetUserById обрабатывает AJAX запрос для получения данных пользователя по ID.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//  1. Проверяет, что пользователь - администратор
//  2. Получает ID пользователя из параметров запроса
//  3. Загружает пользователя из БД вместе с группами
//  4. Форматирует ответ в формате PHP (lowercase поля)
//
// Формат запроса:
//
//	GET /users/ajax_getid_user?id=1
//	или
//	POST /users/ajax_getid_user с параметром id
//
// Формат ответа:
//
//	{
//	  "error": false,
//	  "success": true,
//	  "data": {
//	    "id": "1",
//	    "login": "admin",
//	    "groups": "1,2",
//	    "status": "enable",
//	    "last_name": "Doe",
//	    "name": "John",
//	    "email": "admin@example.com"
//	  }
//	}
func (u *UserController) AjaxGetUserById(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"success": false,
		})
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access deny",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 2: Получение ID пользователя из параметров
	// ============================================
	// Поддерживаем оба формата: GET (query) и POST (form)
	var userIDStr string
	if c.Request.Method == "GET" {
		userIDStr = c.Query("id")
	} else {
		userIDStr = c.PostForm("id")
	}

	if userIDStr == "" {
		c.JSON(http.StatusOK, gin.H{
			"error":   "User ID is required",
			"success": false,
		})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Invalid user ID",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 3: Получение пользователя из БД
	// ============================================
	userRepo := repositories.NewUserRepository()
	targetUser, err := userRepo.FindByID(userID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", userID).
			Str("client_ip", c.ClientIP()).
			Msg("Failed to get user by ID")

		c.JSON(http.StatusOK, gin.H{
			"error":   "DB ERROR: Error select user",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 4: Загрузка групп пользователя
	// ============================================
	groups, err := userRepo.FindGroupsByUserID(userID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", userID).
			Msg("Failed to get user groups")
		groups = []int{} // Продолжаем с пустым списком групп
	}

	// Конвертируем группы в строку через запятую (как в PHP)
	groupsStr := ""
	if len(groups) > 0 {
		groupStrs := make([]string, len(groups))
		for i, gid := range groups {
			groupStrs[i] = strconv.Itoa(gid)
		}
		groupsStr = strings.Join(groupStrs, ",")
	}

	// ============================================
	// ШАГ 5: Форматирование ответа (как в PHP - lowercase поля)
	// ============================================
	status := utils.BoolToStatus(targetUser.Active)

	responseData := gin.H{
		"id":        strconv.Itoa(targetUser.ID),
		"login":     targetUser.Login,
		"groups":    groupsStr,
		"status":    status,
		"last_name": targetUser.LastName,
		"name":      targetUser.Name,
		"email":     targetUser.Email,
		"timezone":  targetUser.Timezone,
	}

	if targetUser.Timezone == "" {
		responseData["timezone"] = "UTC"
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
		"data":    responseData,
	})
}

// AjaxCreateUser обрабатывает AJAX запрос для создания нового пользователя.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//  1. Проверяет, что пользователь - администратор
//  2. Парсит и валидирует все поля формы
//  3. Проверяет уникальность логина и email
//  4. Валидирует пароль
//  5. Создаёт пользователя в БД
//
// Формат запроса:
//
//	POST /users/ajax_create_user
//	Параметры:
//	  - create_user_login: логин
//	  - create_user_password: пароль
//	  - create_user_password_confirm: подтверждение пароля
//	  - create_user_email: email
//	  - create_user_groups: группы через запятую (например, "1,2")
//	  - create_user_status: статус ("enable" или "disable")
//	  - create_user_name: имя
//	  - create_user_last_name: фамилия
//
// Формат ответа:
//
//	{
//	  "error": false,
//	  "success": true
//	}
//	или
//	{
//	  "error": "Error message",
//	  "success": false
//	}
func (u *UserController) AjaxCreateUser(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Unauthorized",
			"success": false,
		})
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Access deny",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 2: Парсинг параметров формы
	// ============================================
	login := strings.TrimSpace(c.PostForm("create_user_login"))
	password := c.PostForm("create_user_password")
	passwordConfirm := c.PostForm("create_user_password_confirm")
	email := strings.TrimSpace(c.PostForm("create_user_email"))
	groupsStr := c.PostForm("create_user_groups")
	status := c.PostForm("create_user_status")
	name := strings.TrimSpace(c.PostForm("create_user_name"))
	lastName := strings.TrimSpace(c.PostForm("create_user_last_name"))
	timezone := strings.TrimSpace(c.PostForm("create_user_timezone"))
	if timezone == "" {
		timezone = "UTC"
	}

	// ============================================
	// ШАГ 3: Валидация обязательных полей
	// ============================================
	var errorMsg string

	if login == "" {
		errorMsg = "Field \"Login\" is empty"
	} else if password == "" {
		errorMsg = "Field \"Password\" is empty"
	} else if passwordConfirm == "" {
		errorMsg = "Field \"Password Confirm\" is empty"
	} else if email == "" {
		errorMsg = "Field \"Email\" is empty"
	} else if groupsStr == "" {
		errorMsg = "Field \"Groups\" is empty"
	} else if status == "" {
		errorMsg = "Field \"Status\" is empty"
	} else if name == "" {
		errorMsg = "Field \"Name\" is empty"
	} else if lastName == "" {
		errorMsg = "Field \"Last Name\" is empty"
	}

	if errorMsg != "" {
		c.JSON(http.StatusOK, gin.H{
			"error":   errorMsg,
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 4: Валидация пароля
	// ============================================
	if err := utils.PasswordValidateWithConfirm(password, passwordConfirm); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 5: Валидация статуса
	// ============================================
	if err := utils.ValidateStatus(status); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 6: Валидация email
	// ============================================
	if err := utils.ValidateEmail(email); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 6.3: Валидация часового пояса
	// ============================================
	if _, err := time.LoadLocation(timezone); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "timezone failed",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 6.1: Создание репозитория для проверки уникальности
	// ============================================
	userRepo := repositories.NewUserRepository()

	// ============================================
	// ШАГ 6.2: Проверка уникальности email
	// ============================================
	existsEmail, err := userRepo.ExistsByEmail(email)
	if err != nil {
		logger.Error().
			Err(err).
			Str("email", email).
			Msg("Failed to check email uniqueness")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Database error",
			"success": false,
		})
		return
	}
	if existsEmail {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Email already exist",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 7: Проверка уникальности логина
	// ============================================
	existsLogin, err := userRepo.ExistsByLogin(login)
	if err != nil {
		logger.Error().
			Err(err).
			Str("login", login).
			Msg("Failed to check login uniqueness")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Database error",
			"success": false,
		})
		return
	}
	if existsLogin {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Login already exist",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 8: Парсинг и валидация групп
	// ============================================
	groupIDs, err := utils.ParseGroupIDs(groupsStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// Проверяем, что все группы существуют
	if err := utils.ValidateGroupIDs(groupIDs); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Failed groups define",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 9: Хеширование пароля
	// ============================================
	hashedPassword, err := utils.PasswordHash(password)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to hash password")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Failed to hash password",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 10: Создание пользователя
	// ============================================
	newUser := &models.User{
		Login:    login,
		Password: hashedPassword,
		Email:    email,
		Active:   utils.StatusToBool(status),
		Name:     name,
		LastName: lastName,
		Timezone: timezone,
	}

	createdUserID, err := userRepo.Create(newUser, groupIDs, user.ID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("login", login).
			Str("client_ip", c.ClientIP()).
			Msg("Failed to create user")

		c.JSON(http.StatusOK, gin.H{
			"error":   "Error create user",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 11: Логирование успешного создания
	// ============================================
	logger.Info().
		Int("created_user_id", createdUserID).
		Str("login", login).
		Int("creator_user_id", user.ID).
		Str("client_ip", c.ClientIP()).
		Msg("User created successfully")

	// ============================================
	// ШАГ 12: Возврат успешного ответа
	// ============================================
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
	})
}

// AjaxEditUser обрабатывает AJAX запрос для редактирования пользователя.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//  1. Проверяет, что пользователь - администратор
//  2. Парсит и валидирует все поля формы
//  3. Проверяет уникальность логина и email (исключая текущего пользователя)
//  4. Валидирует пароль (если указан)
//  5. Обновляет пользователя в БД
//
// Формат запроса:
//
//	POST /users/ajax_edit_user
//	Параметры:
//	  - edit_user_id: ID пользователя
//	  - edit_user_login: логин
//	  - edit_user_password: пароль (опционально, если пустой - не обновляется)
//	  - edit_user_password_confirm: подтверждение пароля (если пароль указан)
//	  - edit_user_email: email
//	  - edit_user_groups: группы через запятую (например, "1,2")
//	  - edit_user_status: статус ("enable" или "disable")
//	  - edit_user_name: имя
//	  - edit_user_last_name: фамилия
//
// Формат ответа:
//
//	{
//	  "error": false,
//	  "success": true
//	}
//	или
//	{
//	  "error": "Error message",
//	  "success": false
//	}
func (u *UserController) AjaxEditUser(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Unauthorized",
			"success": false,
		})
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Access deny",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 2: Парсинг параметров формы
	// ============================================
	userIDStr := c.PostForm("edit_user_id")
	login := strings.TrimSpace(c.PostForm("edit_user_login"))
	password := c.PostForm("edit_user_password")
	passwordConfirm := c.PostForm("edit_user_password_confirm")
	email := strings.TrimSpace(c.PostForm("edit_user_email"))
	groupsStr := c.PostForm("edit_user_groups")
	status := c.PostForm("edit_user_status")
	name := strings.TrimSpace(c.PostForm("edit_user_name"))
	lastName := strings.TrimSpace(c.PostForm("edit_user_last_name"))
	timezone := strings.TrimSpace(c.PostForm("edit_user_timezone"))
	if timezone == "" {
		timezone = "UTC"
	}

	// ============================================
	// ШАГ 3: Валидация обязательных полей
	// ============================================
	var errorMsg string

	if userIDStr == "" {
		errorMsg = "Field \"User ID\" is empty"
	} else if login == "" {
		errorMsg = "Field \"Login\" is empty"
	} else if email == "" {
		errorMsg = "Field \"Email\" is empty"
	} else if groupsStr == "" {
		errorMsg = "Field \"Groups\" is empty"
	} else if status == "" {
		errorMsg = "Field \"Status\" is empty"
	} else if name == "" {
		errorMsg = "Field \"Name\" is empty"
	} else if lastName == "" {
		errorMsg = "Field \"Last Name\" is empty"
	}

	// Если пароль указан, проверяем и подтверждение
	if password != "" && passwordConfirm == "" {
		errorMsg = "Field \"Password Confirm\" is empty"
	}

	if errorMsg != "" {
		c.JSON(http.StatusOK, gin.H{
			"error":   errorMsg,
			"success": false,
		})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Invalid user ID",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 4: Проверка уникальности логина (исключая текущего пользователя)
	// ============================================
	userRepo := repositories.NewUserRepository()
	existsLogin, err := userRepo.ExistsByLoginExcludingID(login, userID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("login", login).
			Int("user_id", userID).
			Msg("Failed to check login uniqueness")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Database error",
			"success": false,
		})
		return
	}
	if existsLogin {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Login already exist",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 5: Валидация пароля (если указан)
	// ============================================
	updatePassword := password != ""
	if updatePassword {
		if err := utils.PasswordValidateWithConfirm(password, passwordConfirm); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}
	}

	// ============================================
	// ШАГ 6: Валидация статуса
	// ============================================
	if err := utils.ValidateStatus(status); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 7: Валидация email
	// ============================================
	if err := utils.ValidateEmail(email); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 7.2: Валидация часового пояса
	// ============================================
	if _, err := time.LoadLocation(timezone); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "timezone failed",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 7.1: Проверка уникальности email (исключая текущего пользователя)
	// ============================================
	existsEmail, err := userRepo.ExistsByEmailExcludingID(email, userID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("email", email).
			Int("user_id", userID).
			Msg("Failed to check email uniqueness")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Database error",
			"success": false,
		})
		return
	}
	if existsEmail {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Email already exist",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 8: Парсинг и валидация групп
	// ============================================
	groupIDs, err := utils.ParseGroupIDs(groupsStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// Проверяем, что все группы существуют
	if err := utils.ValidateGroupIDs(groupIDs); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Failed groups define",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 9: Получение существующего пользователя
	// ============================================
	existingUser, err := userRepo.FindByID(userID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", userID).
			Msg("Failed to find user for update")
		c.JSON(http.StatusOK, gin.H{
			"error":   "User not found",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 10: Хеширование пароля (если указан)
	// ============================================
	if updatePassword {
		hashedPassword, err := utils.PasswordHash(password)
		if err != nil {
			logger.Error().
				Err(err).
				Msg("Failed to hash password")
			c.JSON(http.StatusOK, gin.H{
				"error":   "Failed to hash password",
				"success": false,
			})
			return
		}
		existingUser.Password = hashedPassword
	}

	// ============================================
	// ШАГ 11: Обновление данных пользователя
	// ============================================
	existingUser.Login = login
	existingUser.Email = email
	existingUser.Active = utils.StatusToBool(status)
	existingUser.Name = name
	existingUser.LastName = lastName
	existingUser.Timezone = timezone

	err = userRepo.Update(existingUser, groupIDs, updatePassword, user.ID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", userID).
			Str("client_ip", c.ClientIP()).
			Msg("Failed to update user")

		c.JSON(http.StatusOK, gin.H{
			"error":   "Error edit user",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 12: Логирование успешного обновления
	// ============================================
	logger.Info().
		Int("updated_user_id", userID).
		Str("login", login).
		Int("editor_user_id", user.ID).
		Str("client_ip", c.ClientIP()).
		Msg("User updated successfully")

	// ============================================
	// ШАГ 13: Возврат успешного ответа
	// ============================================
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
	})
}
