package assets

import (
	"github.com/gofiber/fiber/v2"
	"log"
)

func (s Assets) GetFile() fiber.Handler {
	return func(ctx *fiber.Ctx) error {

		if ctx.Params("namespace") == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"ok": false, "err": "set namespace"})
		}

		path := ctx.Params("namespace") + "/" + ctx.Params("key") + "/" + ctx.Params("filename")
		log.Println(path)
		exist := s.KeyExists(ctx.Context(), path)
		if !exist {
			return ctx.SendStatus(fiber.StatusNotFound)
		}
		url, err := s.GetObjectURL(path)
		log.Println(url)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"ok": false, "err": err})
		}
		return ctx.Redirect(url.URL, fiber.StatusPermanentRedirect)
	}
}
