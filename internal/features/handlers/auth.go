package handlers

import "github.com/gofiber/fiber/v2"

func (h *Handler) SetUpRoutes(app *fiber.App) {

	apiV1 := app.Group("/api/v1")

	apiV1.Post("login", h.Auth.LoginHandler())

	apiV1.Post("register", h.Auth.RegisterHandler())

	apiV1.Post("otp/send", h.Auth.SendOTPHandler())

	apiV1.Post("otp/verify", h.Auth.VerifyOtpHandler())

	apiV1.Post("reset/password", h.Auth.ResetPasswordHandler())

	apiV1.Get("logout", MustAuth(h.JwtSecret), h.Auth.LogoutHandler())

	apiV1.Post("refresh", h.Auth.RefreshTokenHandler())

	apiV1.Get("test-login", MustAuth(h.JwtSecret), func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})
}
