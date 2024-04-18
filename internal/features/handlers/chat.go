package handlers

import (
	"github.com/NuEventTeam/events/internal/features/chat"
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) SetUpChatRoutes(router *fiber.App) {
	apiV1 := router.Group("/api/v1")

	apiV1.Get("/event/chat/preview/:eventId", MustAuth(h.JwtSecret), chat.GetChats(h.DB))

	apiV1.Get("/event/chat/messages/:eventId", MustAuth(h.JwtSecret), chat.GetChatMessages(h.DB))

}
