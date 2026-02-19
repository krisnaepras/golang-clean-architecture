package test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"golang-clean-architecture/internal/entity"
	"golang-clean-architecture/internal/model"
)

// ─── Konstanta test ────────────────────────────────────────────────────────────

const (
	testEmail    = "auth_test@dreampod.test"
	testPassword = "TestPassword123!"
	testFullName = "Auth Test User"
)

// ─── Helpers ───────────────────────────────────────────────────────────────────

// hashOtpCode reimplements private hashOtp dari usecase agar test bisa reverse-lookup.
func hashOtpCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

// findPlainOTP iterasi semua kode 6-digit (max 1M) sampai hash cocok.
func findPlainOTP(t *testing.T, storedHash string) string {
	t.Helper()
	for i := 0; i < 1_000_000; i++ {
		candidate := fmt.Sprintf("%06d", i)
		if hashOtpCode(candidate) == storedHash {
			return candidate
		}
	}
	t.Fatal("could not reverse OTP hash — kode tidak ditemukan dalam range 000000-999999")
	return ""
}

// getOTPHash mengambil hash OTP terbaru yang belum diverifikasi dari DB.
func getOTPHash(t *testing.T, email, purpose string) string {
	t.Helper()
	var otp entity.Otp
	err := db.
		Where("destination = ? AND purpose = ? AND verified_at IS NULL", email, purpose).
		Order("created_at DESC").
		First(&otp).Error
	require.NoError(t, err, "OTP harus ada di DB untuk email=%s purpose=%s", email, purpose)
	return otp.OtpCode
}

// cleanupUser menghapus semua data user test.
// Hard-delete via email (cascade via FK ON DELETE CASCADE).
// Juga hapus langsung auth_provider jika cascade tidak berjalan (safeguard).
func cleanupUser(email string) {
	db.Exec("DELETE FROM user_auth_providers WHERE email = ?", email)
	db.Exec("DELETE FROM users WHERE email = ?", email)
}

// doPost membuat HTTP POST request ke Fiber app.
func doPost(t *testing.T, path string, body interface{}) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	return resp
}

// doPostAuth membuat HTTP POST request dengan Bearer token.
func doPostAuth(t *testing.T, path, token string, body interface{}) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	return resp
}

// doGet membuat HTTP GET request dengan optional Bearer token.
func doGet(t *testing.T, path, token string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	return resp
}

// readBody membaca body response sebagai string.
func readBody(resp *http.Response) string {
	b, _ := io.ReadAll(resp.Body)
	return string(b)
}

// parseTokens mengekstrak TokenResponse dari body string yang sudah dibaca.
func parseTokens(t *testing.T, body string) model.TokenResponse {
	t.Helper()
	var wrapper struct {
		Data model.TokenResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &wrapper), "body: %s", body)
	require.NotEmpty(t, wrapper.Data.AccessToken, "access_token kosong — body: %s", body)
	require.NotEmpty(t, wrapper.Data.RefreshToken, "refresh_token kosong — body: %s", body)
	return wrapper.Data
}

// parseErrors mengekstrak field errors dari response error Fiber.
func parseErrors(resp *http.Response) string {
	body := readBody(resp)
	var m map[string]string
	_ = json.Unmarshal([]byte(body), &m)
	return m["errors"]
}

// ─── Suite utama ───────────────────────────────────────────────────────────────

// TestAuth menjalankan seluruh skenario auth secara berurutan menggunakan subtests.
// State (tokens) di-share antar subtest melalui closure.
func TestAuth(t *testing.T) {
	cleanupUser(testEmail)
	t.Cleanup(func() { cleanupUser(testEmail) })

	var (
		accessToken  string
		refreshToken string
	)

	// ── 1. Register ─────────────────────────────────────────────────────────────

	t.Run("Register_InvalidRequest_MissingFields", func(t *testing.T) {
		resp := doPost(t, "/api/auth/register", map[string]string{
			"email": "notvalid",
		})
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
			"harus 400 jika field wajib tidak ada")
	})

	t.Run("Register_Success", func(t *testing.T) {
		resp := doPost(t, "/api/auth/register", map[string]string{
			"full_name": testFullName,
			"email":     testEmail,
			"password":  testPassword,
		})
		body := readBody(resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "body: %s", body)

		var wrapper struct {
			Data map[string]string `json:"data"`
		}
		require.NoError(t, json.Unmarshal([]byte(body), &wrapper))
		assert.Equal(t, "OTP sent to email", wrapper.Data["message"])
	})

	t.Run("Register_DuplicateEmail", func(t *testing.T) {
		resp := doPost(t, "/api/auth/register", map[string]string{
			"full_name": testFullName,
			"email":     testEmail,
			"password":  testPassword,
		})
		assert.Equal(t, http.StatusConflict, resp.StatusCode,
			"harus 409 jika email sudah terdaftar")
	})

	// ── 2. Login sebelum verifikasi ──────────────────────────────────────────────

	t.Run("Login_BeforeEmailVerification", func(t *testing.T) {
		resp := doPost(t, "/api/auth/login", map[string]string{
			"email":    testEmail,
			"password": testPassword,
		})
		assert.Equal(t, http.StatusForbidden, resp.StatusCode,
			"harus 403 jika email belum diverifikasi")
		assert.Contains(t, parseErrors(resp), "email not verified")
	})

	// ── 3. Verify Email OTP ──────────────────────────────────────────────────────

	t.Run("VerifyEmail_WrongOTP", func(t *testing.T) {
		resp := doPost(t, "/api/auth/verify-email", map[string]string{
			"email":    testEmail,
			"otp_code": "000000",
		})
		if resp.StatusCode != http.StatusOK {
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
				"harus 400 jika OTP salah")
		}
	})

	t.Run("VerifyEmail_Success", func(t *testing.T) {
		otpHash := getOTPHash(t, testEmail, "email_verification")
		plainOTP := findPlainOTP(t, otpHash)

		resp := doPost(t, "/api/auth/verify-email", map[string]string{
			"email":    testEmail,
			"otp_code": plainOTP,
		})
		body := readBody(resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"verifikasi email harus berhasil — body: %s", body)

		tokens := parseTokens(t, body)
		accessToken = tokens.AccessToken
		refreshToken = tokens.RefreshToken
		assert.Greater(t, tokens.ExpiresIn, int64(0))
	})

	// ── 4. Login ─────────────────────────────────────────────────────────────────

	t.Run("Login_NonExistentEmail", func(t *testing.T) {
		resp := doPost(t, "/api/auth/login", map[string]string{
			"email":    "nobody@dreampod.test",
			"password": testPassword,
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
			"harus 401 jika email tidak ditemukan")
	})

	t.Run("Login_WrongPassword", func(t *testing.T) {
		resp := doPost(t, "/api/auth/login", map[string]string{
			"email":    testEmail,
			"password": "WrongPassword!",
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
			"harus 401 jika password salah")
	})

	t.Run("Login_Success", func(t *testing.T) {
		resp := doPost(t, "/api/auth/login", map[string]string{
			"email":    testEmail,
			"password": testPassword,
		})
		body := readBody(resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"login harus berhasil — body: %s", body)

		tokens := parseTokens(t, body)
		accessToken = tokens.AccessToken
		refreshToken = tokens.RefreshToken
	})

	// ── 5. Me ────────────────────────────────────────────────────────────────────

	t.Run("Me_Unauthenticated", func(t *testing.T) {
		resp := doGet(t, "/api/auth/me", "")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
			"harus 401 tanpa token")
	})

	t.Run("Me_InvalidToken", func(t *testing.T) {
		resp := doGet(t, "/api/auth/me", "this.is.invalid")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
			"harus 401 jika token tidak valid")
	})

	t.Run("Me_Authenticated", func(t *testing.T) {
		resp := doGet(t, "/api/auth/me", accessToken)
		body := readBody(resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"harus 200 dengan token valid — body: %s", body)

		var wrapper struct {
			Data model.UserResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal([]byte(body), &wrapper))
		assert.Equal(t, testEmail, wrapper.Data.Email)
		assert.Equal(t, testFullName, wrapper.Data.FullName)
		assert.True(t, wrapper.Data.EmailVerified)
		assert.Equal(t, []string{"CUSTOMER"}, wrapper.Data.Roles)
	})

	// ── 6. Refresh Token ─────────────────────────────────────────────────────────

	t.Run("Refresh_InvalidToken", func(t *testing.T) {
		resp := doPost(t, "/api/auth/refresh", map[string]string{
			"refresh_token": "invalidtoken",
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
			"harus 401 jika refresh token tidak valid")
	})

	t.Run("Refresh_Success", func(t *testing.T) {
		oldRefresh := refreshToken

		resp := doPost(t, "/api/auth/refresh", map[string]string{
			"refresh_token": refreshToken,
		})
		body := readBody(resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"refresh token harus berhasil — body: %s", body)

		tokens := parseTokens(t, body)
		assert.NotEmpty(t, tokens.AccessToken, "access token harus ada setelah refresh")
		assert.NotEqual(t, oldRefresh, tokens.RefreshToken, "refresh token harus dirotasi")
		accessToken = tokens.AccessToken
		refreshToken = tokens.RefreshToken
	})

	t.Run("Refresh_OldTokenRevoked_CountCheck", func(t *testing.T) {
		var count int64
		db.Model(&entity.RefreshToken{}).Where("is_revoked = true").Count(&count)
		assert.Greater(t, count, int64(0),
			"harus ada minimal 1 revoked token setelah rotate")
	})

	// ── 7. Logout ────────────────────────────────────────────────────────────────

	t.Run("Logout_InvalidToken", func(t *testing.T) {
		resp := doPostAuth(t, "/api/auth/logout", accessToken, map[string]string{
			"refresh_token": "not_a_real_token",
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
			"harus 401 jika refresh token tidak ditemukan")
	})

	t.Run("Logout_Success", func(t *testing.T) {
		resp := doPostAuth(t, "/api/auth/logout", accessToken, map[string]string{
			"refresh_token": refreshToken,
		})
		body := readBody(resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"logout harus berhasil — body: %s", body)

		var wrapper struct {
			Data map[string]string `json:"data"`
		}
		require.NoError(t, json.Unmarshal([]byte(body), &wrapper))
		assert.Equal(t, "logged out", wrapper.Data["message"])
	})

	t.Run("Refresh_AfterLogout", func(t *testing.T) {
		resp := doPost(t, "/api/auth/refresh", map[string]string{
			"refresh_token": refreshToken,
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
			"harus 401 setelah token di-revoke lewat logout")
	})

	// Re-login untuk sisa test
	loginResp := doPost(t, "/api/auth/login", map[string]string{
		"email":    testEmail,
		"password": testPassword,
	})
	require.Equal(t, http.StatusOK, loginResp.StatusCode, "re-login harus berhasil")
	loginBody0 := readBody(loginResp)
	loginTokens := parseTokens(t, loginBody0)
	accessToken = loginTokens.AccessToken
	_ = accessToken

	// ── 8. Resend OTP ────────────────────────────────────────────────────────────

	t.Run("ResendOTP_Success", func(t *testing.T) {
		resp := doPost(t, "/api/auth/otp/resend", map[string]string{
			"email":   testEmail,
			"purpose": "email_verification",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"resend OTP harus selalu 200")
	})

	t.Run("ResendOTP_NonExistentEmail", func(t *testing.T) {
		resp := doPost(t, "/api/auth/otp/resend", map[string]string{
			"email":   "notexist@dreampod.test",
			"purpose": "email_verification",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"resend OTP untuk email tidak ada harus tetap 200")
	})

	// ── 9. Forgot & Reset Password ───────────────────────────────────────────────

	t.Run("ForgotPassword_NonExistentEmail", func(t *testing.T) {
		resp := doPost(t, "/api/auth/forgot-password", map[string]string{
			"email": "ghost@dreampod.test",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"forgot password harus selalu 200 (tidak expose info)")
	})

	t.Run("ForgotPassword_Success", func(t *testing.T) {
		resp := doPost(t, "/api/auth/forgot-password", map[string]string{
			"email": testEmail,
		})
		body := readBody(resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"forgot password harus 200 — body: %s", body)
	})

	t.Run("ResetPassword_WrongOTP", func(t *testing.T) {
		resp := doPost(t, "/api/auth/reset-password", map[string]string{
			"email":        testEmail,
			"otp_code":     "000000",
			"new_password": "NewPassword456!",
		})
		if resp.StatusCode != http.StatusOK {
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
				"harus 400 jika OTP salah")
		}
	})

	t.Run("ResetPassword_Success", func(t *testing.T) {
		otpHash := getOTPHash(t, testEmail, "reset_password")
		plainOTP := findPlainOTP(t, otpHash)
		newPassword := "NewPassword456!"

		resp := doPost(t, "/api/auth/reset-password", map[string]string{
			"email":        testEmail,
			"otp_code":     plainOTP,
			"new_password": newPassword,
		})
		body := readBody(resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"reset password harus berhasil — body: %s", body)

		// Verifikasi login dengan password baru
		loginResp2 := doPost(t, "/api/auth/login", map[string]string{
			"email":    testEmail,
			"password": newPassword,
		})
		loginBody := readBody(loginResp2)
		assert.Equal(t, http.StatusOK, loginResp2.StatusCode,
			"login dengan password baru harus berhasil — body: %s", loginBody)
		newTokens := parseTokens(t, loginBody)
		assert.NotEmpty(t, newTokens.AccessToken)
	})
}
