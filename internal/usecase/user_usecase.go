package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"golang-clean-architecture/internal/entity"
	"golang-clean-architecture/internal/model"
	"golang-clean-architecture/internal/model/converter"
	"golang-clean-architecture/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 30 * 24 * time.Hour
	otpTTL          = 10 * time.Minute
)

type UserUseCase struct {
	DB                     *gorm.DB
	Log                    *logrus.Logger
	Validate               *validator.Validate
	JwtSecret              string
	UserRepository         *repository.UserRepository
	AuthProviderRepository *repository.AuthProviderRepository
	RefreshTokenRepository *repository.RefreshTokenRepository
	OtpRepository          *repository.OtpRepository
	OauthStateRepository   *repository.OauthStateRepository
}

func NewUserUseCase(
	db *gorm.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	jwtSecret string,
	userRepo *repository.UserRepository,
	authProviderRepo *repository.AuthProviderRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	otpRepo *repository.OtpRepository,
	oauthStateRepo *repository.OauthStateRepository,
) *UserUseCase {
	return &UserUseCase{
		DB:                     db,
		Log:                    logger,
		Validate:               validate,
		JwtSecret:              jwtSecret,
		UserRepository:         userRepo,
		AuthProviderRepository: authProviderRepo,
		RefreshTokenRepository: refreshTokenRepo,
		OtpRepository:          otpRepo,
		OauthStateRepository:   oauthStateRepo,
	}
}

// ── JWT helpers ────────────────────────────────────────────────────────────────

func (c *UserUseCase) generateAccessToken(userID string, roles []string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"roles": roles,
		"exp":   time.Now().Add(accessTokenTTL).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.JwtSecret))
}

func (c *UserUseCase) parseAccessToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(c.JwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fiber.ErrUnauthorized
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fiber.ErrUnauthorized
	}
	return claims, nil
}

func generateRefreshToken() (plain, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	plain = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(plain))
	hash = hex.EncodeToString(sum[:])
	return
}

func hashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

func generateOtpCode() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	n := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	return fmt.Sprintf("%06d", n%1_000_000)
}

func hashOtp(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func (c *UserUseCase) issueTokenPair(db *gorm.DB, userID string, roles []string, deviceInfo, ip string) (*model.TokenResponse, error) {
	accessToken, err := c.generateAccessToken(userID, roles)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	plain, hash, err := generateRefreshToken()
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	rt := &entity.RefreshToken{
		ID:         uuid.New().String(),
		UserID:     userID,
		TokenHash:  hash,
		DeviceInfo: deviceInfo,
		IPAddress:  ip,
		IsRevoked:  false,
		ExpiredAt:  time.Now().Add(refreshTokenTTL),
	}
	if err := c.RefreshTokenRepository.Create(db, rt); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return &model.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: plain,
		ExpiresIn:    int64(accessTokenTTL.Seconds()),
	}, nil
}

// ── Verify (dipakai middleware) ────────────────────────────────────────────────

func (c *UserUseCase) Verify(ctx context.Context, request *model.VerifyUserRequest) (*model.Auth, error) {
	if err := c.Validate.Struct(request); err != nil {
		return nil, fiber.ErrUnauthorized
	}

	// strip "Bearer " prefix if present
	tokenStr := request.Token
	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}

	claims, err := c.parseAccessToken(tokenStr)
	if err != nil {
		return nil, fiber.ErrUnauthorized
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return nil, fiber.ErrUnauthorized
	}

	rawRoles, _ := claims["roles"].([]interface{})
	roles := make([]string, 0, len(rawRoles))
	for _, r := range rawRoles {
		if s, ok := r.(string); ok {
			roles = append(roles, s)
		}
	}

	return &model.Auth{ID: userID, Roles: roles}, nil
}

// ── Register ───────────────────────────────────────────────────────────────────

func (c *UserUseCase) Register(ctx context.Context, request *model.RegisterRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("Invalid register request: %+v", err)
		return fiber.ErrBadRequest
	}

	// cek email sudah terdaftar
	existing := new(entity.User)
	if err := c.UserRepository.FindByEmail(tx, existing, request.Email); err == nil {
		return fiber.ErrConflict
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	user := &entity.User{
		ID:            uuid.New().String(),
		FullName:      request.FullName,
		Email:         request.Email,
		IsActive:      true,
		EmailVerified: false,
	}
	if err := c.UserRepository.Create(tx, user); err != nil {
		c.Log.Warnf("Failed create user: %+v", err)
		return fiber.ErrInternalServerError
	}

	provider := &entity.UserAuthProvider{
		ID:             uuid.New().String(),
		UserID:         user.ID,
		Provider:       "email",
		ProviderUserID: request.Email, // unique key per email for email-auth provider
		Email:          request.Email,
		PasswordHash:   string(passwordHash),
	}
	if err := c.AuthProviderRepository.Create(tx, provider); err != nil {
		c.Log.Warnf("Failed create auth provider: %+v", err)
		return fiber.ErrInternalServerError
	}

	// assign default role CUSTOMER
	if err := tx.Exec("INSERT INTO user_roles (user_id, role_id) SELECT ?, id FROM roles WHERE code = 'CUSTOMER'", user.ID).Error; err != nil {
		c.Log.Warnf("Failed assign role: %+v", err)
		return fiber.ErrInternalServerError
	}

	otpCode := generateOtpCode()
	otpExpiry := time.Now().Add(otpTTL)
	otp := &entity.Otp{
		ID:          uuid.New().String(),
		UserID:      user.ID,
		Destination: request.Email,
		OtpCode:     hashOtp(otpCode),
		Purpose:     "email_verification",
		ExpiredAt:   &otpExpiry,
	}
	if err := c.OtpRepository.Create(tx, otp); err != nil {
		c.Log.Warnf("Failed create OTP: %+v", err)
		return fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.ErrInternalServerError
	}

	// TODO: kirim email OTP ke request.Email dengan kode otpCode
	c.Log.Infof("OTP for %s: %s (purpose: email_verification)", request.Email, otpCode)
	return nil
}

// ── Verify Email OTP ───────────────────────────────────────────────────────────

func (c *UserUseCase) VerifyEmailOtp(ctx context.Context, request *model.VerifyOtpRequest, deviceInfo, ip string) (*model.TokenResponse, error) {
	request.Purpose = "email_verification"
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		return nil, fiber.ErrBadRequest
	}

	otp := new(entity.Otp)
	if err := c.OtpRepository.FindActiveByDestinationAndPurpose(tx, otp, request.Email, "email_verification"); err != nil {
		c.Log.Warnf("OTP not found: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	if otp.OtpCode != hashOtp(request.OtpCode) {
		otp.AttemptCount++
		_ = c.OtpRepository.Update(tx, otp)
		_ = tx.Commit()
		return nil, fiber.ErrBadRequest
	}

	now := time.Now()
	otp.VerifiedAt = &now
	if err := c.OtpRepository.Update(tx, otp); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	user := new(entity.User)
	if err := c.UserRepository.FindById(tx, user, otp.UserID); err != nil {
		return nil, fiber.ErrInternalServerError
	}
	user.EmailVerified = true
	if err := c.UserRepository.Update(tx, user); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	_, roles, err := c.UserRepository.FindWithRoles(tx, user.ID)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	tokens, err := c.issueTokenPair(tx, user.ID, roles, deviceInfo, ip)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}
	return tokens, nil
}

// ── Login ──────────────────────────────────────────────────────────────────────

func (c *UserUseCase) Login(ctx context.Context, request *model.LoginRequest, deviceInfo, ip string) (*model.TokenResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		return nil, fiber.ErrBadRequest
	}

	user := new(entity.User)
	if err := c.UserRepository.FindByEmail(tx, user, request.Email); err != nil {
		return nil, fiber.ErrUnauthorized
	}

	if !user.IsActive {
		return nil, fiber.ErrForbidden
	}

	provider := new(entity.UserAuthProvider)
	if err := c.AuthProviderRepository.FindByUserAndProvider(tx, provider, user.ID, "email"); err != nil {
		return nil, fiber.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(provider.PasswordHash), []byte(request.Password)); err != nil {
		return nil, fiber.ErrUnauthorized
	}

	if !user.EmailVerified {
		return nil, fiber.NewError(fiber.StatusForbidden, "email not verified")
	}

	_, roles, err := c.UserRepository.FindWithRoles(tx, user.ID)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	tokens, err := c.issueTokenPair(tx, user.ID, roles, deviceInfo, ip)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}
	return tokens, nil
}

// ── Refresh Token ──────────────────────────────────────────────────────────────

func (c *UserUseCase) Refresh(ctx context.Context, request *model.RefreshRequest, deviceInfo, ip string) (*model.TokenResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		return nil, fiber.ErrBadRequest
	}

	hash := hashRefreshToken(request.RefreshToken)
	rt := new(entity.RefreshToken)
	if err := c.RefreshTokenRepository.FindByHash(tx, rt, hash); err != nil {
		return nil, fiber.ErrUnauthorized
	}

	if time.Now().After(rt.ExpiredAt) {
		return nil, fiber.ErrUnauthorized
	}

	// rotate: revoke lama
	rt.IsRevoked = true
	if err := c.RefreshTokenRepository.Update(tx, rt); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	_, roles, err := c.UserRepository.FindWithRoles(tx, rt.UserID)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	tokens, err := c.issueTokenPair(tx, rt.UserID, roles, deviceInfo, ip)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}
	return tokens, nil
}

// ── Logout ─────────────────────────────────────────────────────────────────────

func (c *UserUseCase) Logout(ctx context.Context, request *model.LogoutRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		return fiber.ErrBadRequest
	}

	hash := hashRefreshToken(request.RefreshToken)
	rt := new(entity.RefreshToken)
	if err := c.RefreshTokenRepository.FindByHash(tx, rt, hash); err != nil {
		return fiber.ErrUnauthorized
	}

	rt.IsRevoked = true
	if err := c.RefreshTokenRepository.Update(tx, rt); err != nil {
		return fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.ErrInternalServerError
	}
	return nil
}

// ── Me ─────────────────────────────────────────────────────────────────────────

func (c *UserUseCase) Me(ctx context.Context, userID string) (*model.UserResponse, error) {
	db := c.DB.WithContext(ctx)

	user, roles, err := c.UserRepository.FindWithRoles(db, userID)
	if err != nil {
		return nil, fiber.ErrNotFound
	}

	return converter.UserToResponse(user, roles), nil
}

// ── Forgot Password ────────────────────────────────────────────────────────────

func (c *UserUseCase) ForgotPassword(ctx context.Context, request *model.ForgotPasswordRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		return fiber.ErrBadRequest
	}

	user := new(entity.User)
	if err := c.UserRepository.FindByEmail(tx, user, request.Email); err != nil {
		// jangan expose apakah email terdaftar atau tidak
		return nil
	}

	otpCode := generateOtpCode()
	otpExpiry2 := time.Now().Add(otpTTL)
	otp := &entity.Otp{
		ID:          uuid.New().String(),
		UserID:      user.ID,
		Destination: request.Email,
		OtpCode:     hashOtp(otpCode),
		Purpose:     "reset_password",
		ExpiredAt:   &otpExpiry2,
	}
	if err := c.OtpRepository.Create(tx, otp); err != nil {
		return fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.ErrInternalServerError
	}

	// TODO: kirim email OTP ke request.Email
	c.Log.Infof("Reset password OTP for %s: %s", request.Email, otpCode)
	return nil
}

// ── Reset Password ─────────────────────────────────────────────────────────────

func (c *UserUseCase) ResetPassword(ctx context.Context, request *model.ResetPasswordRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		return fiber.ErrBadRequest
	}

	otp := new(entity.Otp)
	if err := c.OtpRepository.FindActiveByDestinationAndPurpose(tx, otp, request.Email, "reset_password"); err != nil {
		return fiber.ErrBadRequest
	}

	if otp.OtpCode != hashOtp(request.OtpCode) {
		otp.AttemptCount++
		_ = c.OtpRepository.Update(tx, otp)
		_ = tx.Commit()
		return fiber.ErrBadRequest
	}

	now2 := time.Now()
	otp.VerifiedAt = &now2
	if err := c.OtpRepository.Update(tx, otp); err != nil {
		return fiber.ErrInternalServerError
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	provider := new(entity.UserAuthProvider)
	if err := c.AuthProviderRepository.FindByUserAndProvider(tx, provider, otp.UserID, "email"); err != nil {
		return fiber.ErrInternalServerError
	}
	provider.PasswordHash = string(newHash)
	if err := c.AuthProviderRepository.Update(tx, provider); err != nil {
		return fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.ErrInternalServerError
	}
	return nil
}

// ── Resend OTP ─────────────────────────────────────────────────────────────────

func (c *UserUseCase) ResendOtp(ctx context.Context, request *model.ResendOtpRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		return fiber.ErrBadRequest
	}

	user := new(entity.User)
	if err := c.UserRepository.FindByEmail(tx, user, request.Email); err != nil {
		return nil // jangan expose
	}

	otpCode := generateOtpCode()
	otpExpiry3 := time.Now().Add(otpTTL)
	otp := &entity.Otp{
		ID:          uuid.New().String(),
		UserID:      user.ID,
		Destination: request.Email,
		OtpCode:     hashOtp(otpCode),
		Purpose:     request.Purpose,
		ExpiredAt:   &otpExpiry3,
	}
	if err := c.OtpRepository.Create(tx, otp); err != nil {
		return fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.ErrInternalServerError
	}

	c.Log.Infof("Resend OTP for %s (purpose: %s): %s", request.Email, request.Purpose, otpCode)
	return nil
}

// ── Google OAuth ───────────────────────────────────────────────────────────────

func (c *UserUseCase) GoogleOAuthURL(ctx context.Context) (string, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	stateVal := uuid.New().String()
	oauthState := &entity.OauthState{
		ID:        uuid.New().String(),
		State:     stateVal,
		Provider:  "google",
		ExpiredAt: time.Now().Add(10 * time.Minute),
	}
	if err := c.OauthStateRepository.Create(tx, oauthState); err != nil {
		return "", fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		return "", fiber.ErrInternalServerError
	}

	// TODO: ganti dengan google OAuth config dari viper
	// Contoh: https://accounts.google.com/o/oauth2/v2/auth?client_id=...
	redirectURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?state=%s&response_type=code&scope=openid+email+profile", stateVal)
	return redirectURL, nil
}

// ── Apple OAuth ────────────────────────────────────────────────────────────────

func (c *UserUseCase) AppleOAuthURL(ctx context.Context) (string, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	stateVal := uuid.New().String()
	oauthState := &entity.OauthState{
		ID:        uuid.New().String(),
		State:     stateVal,
		Provider:  "apple",
		ExpiredAt: time.Now().Add(10 * time.Minute),
	}
	if err := c.OauthStateRepository.Create(tx, oauthState); err != nil {
		return "", fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		return "", fiber.ErrInternalServerError
	}

	// TODO: ganti dengan apple OAuth config dari viper
	redirectURL := fmt.Sprintf("https://appleid.apple.com/auth/authorize?state=%s&response_type=code&scope=name+email&response_mode=form_post", stateVal)
	return redirectURL, nil
}