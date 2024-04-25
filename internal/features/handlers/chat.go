package handlers

import (
	"github.com/NuEventTeam/events/internal/features/chat/chat_features"
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) SetUpChatRoutes(router *fiber.App) {
	apiV1 := router.Group("/api/v1")

	apiV1.Get("/event/chat/preview", MustAuth(h.JwtSecret), chat_features.GetChats(h.DB))

	apiV1.Get("/event/chat/messages/:eventId", MustAuth(h.JwtSecret), chat_features.GetChatMessages(h.DB))

}
