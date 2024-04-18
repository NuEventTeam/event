package handlers

import (
	"github.com/NuEventTeam/events/internal/features/chat"
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) SetUpChatRoutes(router *fiber.App) {

	router.Get("/event/chat/preview/:eventId", MustAuth(h.JwtSecret), chat.GetChats(h.DB))

	router.Get("/event/chat/messages/:eventId", MustAuth(h.JwtSecret), chat.GetChatMessages(h.DB))

}
