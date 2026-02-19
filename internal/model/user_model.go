package model

// ── Register ───────────────────────────────────────────────

type RegisterRequest struct {
	FullName string `json:"full_name" validate:"required,max=100"`
	Email    string `json:"email"     validate:"required,email,max=150"`
	Password string `json:"password"  validate:"required,min=8,max=100"`
}

// ── Login ──────────────────────────────────────────────────

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ── Token ──────────────────────────────────────────────────

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

// ── Refresh ────────────────────────────────────────────────

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ── Logout ─────────────────────────────────────────────────

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ── Me ─────────────────────────────────────────────────────

type UserResponse struct {
	ID            string   `json:"id"`
	FullName      string   `json:"full_name"`
	Email         string   `json:"email"`
	Phone         string   `json:"phone,omitempty"`
	ProfileImage  string   `json:"profile_image,omitempty"`
	IsActive      bool     `json:"is_active"`
	EmailVerified bool     `json:"email_verified"`
	Roles         []string `json:"roles"`
	CreatedAt     int64    `json:"created_at"`
}

// ── OTP ────────────────────────────────────────────────────

type VerifyOtpRequest struct {
	Email   string `json:"email"    validate:"required,email"`
	OtpCode string `json:"otp_code" validate:"required,len=6"`
	Purpose string `json:"-"` // diisi oleh handler
}

type ResendOtpRequest struct {
	Email   string `json:"email"   validate:"required,email"`
	Purpose string `json:"purpose" validate:"required,oneof=email_verification reset_password"`
}

// ── Forgot / Reset password ────────────────────────────────

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"        validate:"required,email"`
	OtpCode     string `json:"otp_code"     validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=100"`
}

// ── OAuth ──────────────────────────────────────────────────

type OAuthCallbackRequest struct {
	Code  string `query:"code"  validate:"required"`
	State string `query:"state" validate:"required"`
}

// ── Verify JWT (internal, dipakai middleware) ──────────────

type VerifyUserRequest struct {
	Token string `validate:"required"`
}
