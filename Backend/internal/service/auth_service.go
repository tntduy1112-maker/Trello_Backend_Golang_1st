package service

import (
	"context"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/internal/dto/request"
	"github.com/codewebkhongkho/trello-agent/internal/dto/response"
	"github.com/codewebkhongkho/trello-agent/internal/repository"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/codewebkhongkho/trello-agent/pkg/cache"
	"github.com/codewebkhongkho/trello-agent/pkg/email"
	"github.com/codewebkhongkho/trello-agent/pkg/hash"
	"github.com/codewebkhongkho/trello-agent/pkg/jwt"
)

const (
	otpExpiresIn        = 15 * time.Minute
	resetTokenExpiresIn = 1 * time.Hour
	maxOTPAttempts      = 5
)

type AuthService struct {
	userRepo         repository.UserRepository
	tokenRepo        repository.TokenRepository
	verificationRepo repository.VerificationRepository
	jwtManager       *jwt.Manager
	emailService     *email.Service
	cache            *cache.RedisClient
	frontendURL      string
}

type AuthServiceConfig struct {
	UserRepo         repository.UserRepository
	TokenRepo        repository.TokenRepository
	VerificationRepo repository.VerificationRepository
	JWTManager       *jwt.Manager
	EmailService     *email.Service
	Cache            *cache.RedisClient
	FrontendURL      string
}

func NewAuthService(cfg AuthServiceConfig) *AuthService {
	return &AuthService{
		userRepo:         cfg.UserRepo,
		tokenRepo:        cfg.TokenRepo,
		verificationRepo: cfg.VerificationRepo,
		jwtManager:       cfg.JWTManager,
		emailService:     cfg.EmailService,
		cache:            cfg.Cache,
		frontendURL:      cfg.FrontendURL,
	}
}

func (s *AuthService) Register(ctx context.Context, req *request.RegisterRequest) (*response.UserResponse, error) {
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if existingUser != nil {
		return nil, apperror.ErrEmailAlreadyExists
	}

	passwordHash, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	user := &domain.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		FullName:     req.FullName,
		IsVerified:   false,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	if err := s.sendVerificationOTP(ctx, user); err != nil {
		return nil, err
	}

	return response.ToUserResponse(user), nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, req *request.VerifyEmailRequest) error {
	attemptsKey := "otp_attempts:" + req.Email
	attempts, _ := s.cache.Incr(ctx, attemptsKey)
	if attempts == 1 {
		_ = s.cache.Expire(ctx, attemptsKey, 15*time.Minute)
	}
	if attempts > maxOTPAttempts {
		return apperror.ErrOTPMaxAttempts
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if user == nil {
		return apperror.ErrUserNotFound
	}
	if user.IsVerified {
		return nil
	}

	verification, err := s.verificationRepo.FindByToken(ctx, req.OTP, domain.VerificationTypeEmail)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if verification == nil || verification.UserID != user.ID {
		return apperror.ErrOTPInvalid
	}
	if !verification.IsValid() {
		if verification.IsExpired() {
			return apperror.ErrOTPExpired
		}
		return apperror.ErrOTPInvalid
	}

	if err := s.verificationRepo.MarkUsed(ctx, verification.ID); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	user.IsVerified = true
	if err := s.userRepo.Update(ctx, user); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	_ = s.cache.Delete(ctx, attemptsKey)
	return nil
}

func (s *AuthService) ResendVerification(ctx context.Context, req *request.ResendVerificationRequest) error {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if user == nil {
		return nil
	}
	if user.IsVerified {
		return nil
	}

	return s.sendVerificationOTP(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, req *request.LoginRequest, deviceInfo, ipAddress string) (*response.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if user == nil {
		return nil, apperror.ErrInvalidCredentials
	}
	if !hash.ComparePassword(user.PasswordHash, req.Password) {
		return nil, apperror.ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, apperror.ErrAccountDisabled
	}
	if !user.IsVerified {
		return nil, apperror.ErrEmailNotVerified
	}

	return s.generateAuthResponse(ctx, user, deviceInfo, ipAddress)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken, deviceInfo, ipAddress string) (*response.AuthResponse, error) {
	tokenHash := hash.SHA256(refreshToken)
	storedToken, err := s.tokenRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if storedToken == nil {
		return nil, apperror.ErrInvalidToken
	}

	if storedToken.IsRevoked {
		_ = s.tokenRepo.RevokeAllUserTokens(ctx, storedToken.UserID)
		_ = s.userRepo.UpdateTokensValidAfter(ctx, storedToken.UserID)
		return nil, apperror.ErrTokenRevoked
	}

	if storedToken.IsExpired() {
		return nil, apperror.ErrTokenExpired
	}

	if err := s.tokenRepo.RevokeToken(ctx, tokenHash); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	user, err := s.userRepo.FindByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if user == nil || !user.IsActive {
		return nil, apperror.ErrAccountDisabled
	}

	return s.generateAuthResponse(ctx, user, deviceInfo, ipAddress)
}

func (s *AuthService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	if accessToken != "" {
		claims, err := s.jwtManager.GetClaimsFromExpiredToken(accessToken)
		if err == nil && claims.ID != "" {
			ttl := time.Until(claims.ExpiresAt.Time)
			if ttl > 0 {
				_ = s.cache.Set(ctx, "blacklist:"+claims.ID, "logout", ttl)
			}
		}
	}

	if refreshToken != "" {
		tokenHash := hash.SHA256(refreshToken)
		_ = s.tokenRepo.RevokeToken(ctx, tokenHash)
	}

	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	if err := s.tokenRepo.RevokeAllUserTokens(ctx, userID); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if err := s.userRepo.UpdateTokensValidAfter(ctx, userID); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	return nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req *request.ForgotPasswordRequest) error {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if user == nil {
		return nil
	}

	token, err := jwt.GenerateResetToken()
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	verification := &domain.EmailVerification{
		UserID:    user.ID,
		Token:     token,
		Type:      domain.VerificationTypePasswordReset,
		ExpiresAt: time.Now().Add(resetTokenExpiresIn),
	}

	if err := s.verificationRepo.Create(ctx, verification); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	if err := s.emailService.SendPasswordResetEmail(user.Email, token, s.frontendURL); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req *request.ResetPasswordRequest) error {
	verification, err := s.verificationRepo.FindByToken(ctx, req.Token, domain.VerificationTypePasswordReset)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if verification == nil || !verification.IsValid() {
		return apperror.ErrInvalidToken
	}

	passwordHash, err := hash.HashPassword(req.Password)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	user, err := s.userRepo.FindByID(ctx, verification.UserID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if user == nil {
		return apperror.ErrUserNotFound
	}

	user.PasswordHash = passwordHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	if err := s.verificationRepo.MarkUsed(ctx, verification.ID); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	_ = s.tokenRepo.RevokeAllUserTokens(ctx, user.ID)
	_ = s.userRepo.UpdateTokensValidAfter(ctx, user.ID)

	return nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID string) (*response.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if user == nil {
		return nil, apperror.ErrUserNotFound
	}
	return response.ToUserResponse(user), nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID string, req *request.UpdateProfileRequest) (*response.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if user == nil {
		return nil, apperror.ErrUserNotFound
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	return response.ToUserResponse(user), nil
}

func (s *AuthService) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	return s.cache.Exists(ctx, "blacklist:"+jti)
}

func (s *AuthService) sendVerificationOTP(ctx context.Context, user *domain.User) error {
	otp, err := jwt.GenerateOTP()
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	_ = s.verificationRepo.DeleteByUserAndType(ctx, user.ID, domain.VerificationTypeEmail)

	verification := &domain.EmailVerification{
		UserID:    user.ID,
		Token:     otp,
		Type:      domain.VerificationTypeEmail,
		ExpiresAt: time.Now().Add(otpExpiresIn),
	}

	if err := s.verificationRepo.Create(ctx, verification); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	if err := s.emailService.SendVerificationEmail(user.Email, otp); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	return nil
}

func (s *AuthService) generateAuthResponse(ctx context.Context, user *domain.User, deviceInfo, ipAddress string) (*response.AuthResponse, error) {
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	refreshTokenHash := hash.SHA256(tokenPair.RefreshToken)
	var deviceInfoPtr, ipAddressPtr *string
	if deviceInfo != "" {
		deviceInfoPtr = &deviceInfo
	}
	if ipAddress != "" {
		ipAddressPtr = &ipAddress
	}

	storedToken := &domain.RefreshToken{
		UserID:     user.ID,
		TokenHash:  refreshTokenHash,
		DeviceInfo: deviceInfoPtr,
		IPAddress:  ipAddressPtr,
		ExpiresAt:  tokenPair.ExpiresAt,
	}

	if err := s.tokenRepo.CreateRefreshToken(ctx, storedToken); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	return &response.AuthResponse{
		User:         response.ToUserResponse(user),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int64(time.Until(tokenPair.ExpiresAt).Seconds()),
	}, nil
}
