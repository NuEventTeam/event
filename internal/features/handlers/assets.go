package handlers

import "github.com/gofiber/fiber/v2"

func (h Handler) SetUpAssetsRoutes(router *fiber.App) {
	router.Get("/get/:namespace/:key/:filename", h.Assets.GetFile())
	router.Static("/static/", "./static")

}
