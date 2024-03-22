package handlers

import "github.com/gofiber/fiber/v2"

func (h *Handler) SetUpUserRoutes(router *fiber.App) {
	apiV1 := router.Group("/api/v1")

	apiV1.Post("/create/user",
		MustAuth(h.JwtSecret),
		h.UserSvc.CreateUserHandler(),
	)

	apiV1.Post("/create/mobile/user",
		MustAuth(h.JwtSecret),
		h.UserSvc.CreateMobileUserHandler(),
	)

	apiV1.Get("/user/:username", h.UserSvc.GetByUsername())

	apiV1.Get("/check-username/:username",
		MustAuth(h.JwtSecret),
		h.UserSvc.GetByUsername(),
	)

	apiV1.Post("/users/friendship/follow/:userId",
		MustAuth(h.JwtSecret),
		h.UserSvc.FollowUser(),
	)

	apiV1.Post("/users/friendship/unfollow/:userId",
		MustAuth(h.JwtSecret),
		h.UserSvc.UnfollowUser(),
	)

}
