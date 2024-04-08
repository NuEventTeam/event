package handlers

import (
	"github.com/NuEventTeam/events/internal/features/event/comments"
	"github.com/NuEventTeam/events/internal/features/event/followers"
	"github.com/NuEventTeam/events/internal/features/event/like"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) SetUpEventRoutes(router *fiber.App) {
	apiV1 := router.Group("/api/v1")
	apiV1.Get("/categories", h.EventSvc.GetAllCategoriesHandler())

	apiV1.Post("/event/create", MustAuth(h.JwtSecret), h.EventSvc.CreateEventHandler())

	apiV1.Get("/event/show/:eventId", h.EventSvc.GetEventByIDHandler())

	apiV1.Put("/event/posts/:eventId",
		MustAuth(h.JwtSecret),
		h.HasPermission(pkg.PermissionUpdate),
		h.EventSvc.UpdateEventHandler(),
	)

	apiV1.Put("/event/image/:eventId",
		MustAuth(h.JwtSecret),
		h.HasPermission(pkg.PermissionUpdate),
		h.EventSvc.AddImage(),
	)

	apiV1.Post("/event/fellowship/follow/:eventId",
		MustAuth(h.JwtSecret),
		followers.FollowEvent(h.DB))

	apiV1.Post("/event/fellowship/unfollow/:eventId",
		MustAuth(h.JwtSecret),
		followers.Unfollow(h.DB))

	apiV1.Post("/event/fellowship/list/:eventId",
		MustAuth(h.JwtSecret),
		followers.ListFollowers(h.DB))

	apiV1.Post("/event/fellowship/exist/:eventId",
		MustAuth(h.JwtSecret),
		followers.CheckIfFollowed(h.DB))

	apiV1.Post("/event/fellowship/search/:eventId",
		MustAuth(h.JwtSecret),
		followers.ListFollowers(h.DB))

	apiV1.Post("/event/comment/add",
		MustAuth(h.JwtSecret),
		comments.AddCommentHandler(h.DB))

	apiV1.Post("/event/comment/fetch",
		comments.FetchCommentHandler(h.DB))

	apiV1.Post("/event/like/:eventId",
		MustAuth(h.JwtSecret),
		like.LikeEvent(h.DB))
}
