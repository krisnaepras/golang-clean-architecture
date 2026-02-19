package middleware

import (
	"golang-clean-architecture/internal/model"
	"golang-clean-architecture/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// NewAuth memverifikasi JWT dan menyimpan auth ke context.
// Gunakan sebagai middleware global atau per-route.
func NewAuth(userUseCase *usecase.UserUseCase) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		request := &model.VerifyUserRequest{Token: ctx.Get("Authorization", "NOT_FOUND")}
		userUseCase.Log.Debugf("Authorization : %s", request.Token)

		auth, err := userUseCase.Verify(ctx.UserContext(), request)
		if err != nil {
			userUseCase.Log.Warnf("Failed find user by token : %+v", err)
			return fiber.ErrUnauthorized
		}

		userUseCase.Log.Debugf("User : %+v", auth.ID)
		ctx.Locals("auth", auth)
		return ctx.Next()
	}
}

// RequireRoles memastikan user yang sudah terautentikasi memiliki
// minimal satu dari role yang diizinkan.
//
// Contoh penggunaan:
//
//	route.Get("/admin", middleware.RequireRoles("ADMIN"))
//	route.Get("/dashboard", middleware.RequireRoles("ADMIN", "MANAGER"))
func RequireRoles(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(ctx *fiber.Ctx) error {
		auth, ok := ctx.Locals("auth").(*model.Auth)
		if !ok || auth == nil {
			return fiber.ErrUnauthorized
		}

		for _, userRole := range auth.Roles {
			if _, found := allowed[userRole]; found {
				return ctx.Next()
			}
		}

		return fiber.ErrForbidden
	}
}

// GetUser mengambil auth dari context Fiber.
func GetUser(ctx *fiber.Ctx) *model.Auth {
	return ctx.Locals("auth").(*model.Auth)
}
