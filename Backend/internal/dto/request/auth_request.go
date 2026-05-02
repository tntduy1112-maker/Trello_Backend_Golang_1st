package request

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	FullName string `json:"full_name" binding:"required,min=2,max=255"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type UpdateProfileRequest struct {
	FullName string `json:"full_name" binding:"omitempty,min=2,max=255"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AcceptInvitationWithPasswordRequest struct {
	FullName string `json:"full_name" binding:"required,min=2,max=255"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}
