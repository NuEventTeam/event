package handlers

import (
	"github.com/NuEventTeam/events/internal/features/search"
	"github.com/NuEventTeam/events/internal/features/user/follow"
	user_profile "github.com/NuEventTeam/events/internal/features/user/profile"
	"github.com/gofiber/fiber/v2"
)

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

	apiV1.Post("users/friendship/check/:userId",
		MustAuth(h.JwtSecret),
		user_follow.CheckFollowed(h.DB),
	)

	apiV1.Post("/users/friendship/follow/:userId",
		MustAuth(h.JwtSecret),
		user_follow.FollowUser(h.DB),
	)

	apiV1.Post("/users/friendship/unfollow/:userId",
		MustAuth(h.JwtSecret),
		user_follow.UnfollowUser(h.DB),
	)
	apiV1.Post("/users/friendship/followed",
		MustAuth(h.JwtSecret),
		user_follow.ListFollowed(h.DB),
	)
	apiV1.Post("/users/friendship/follower",
		MustAuth(h.JwtSecret),
		user_follow.ListFollowers(h.DB),
	)

	apiV1.Post("/users/profile/events/followed",
		MustAuth(h.JwtSecret),
		user_profile.GetFollowedEventsHandler(h.DB),
	)

	apiV1.Post("/users/profile/events/history",
		MustAuth(h.JwtSecret),
		user_profile.GetOldEventsHandler(h.DB),
	)

	apiV1.Post("/users/profile/events/favorite",
		MustAuth(h.JwtSecret),
		user_profile.GetLikedEventsHandler(h.DB),
	)

	apiV1.Post("/users/profile/search/", search.SearchUser(h.DB))

}
