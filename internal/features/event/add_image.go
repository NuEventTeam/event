package event

import (
	"github.com/NuEventTeam/events/internal/features/assets"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/oklog/ulid/v2"
	"log"
	"path"
	"strconv"
	"sync"
)

func (e Event) AddImage() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		form, err := ctx.MultipartForm()
		if err != nil {
			return err
		}

		eventId, err := strconv.ParseInt(ctx.Params("eventId"), 10, 64)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "invalid eventID", err)
		}

		images := make([]*assets.Image, len(form.File["images"]))
		wg := sync.WaitGroup{}

		for i, f := range form.File["images"] {
			wg.Add(1)
			f := f
			go func(index int) {
				defer wg.Done()
				file, err := f.Open()
				if err != nil {
					log.Println("cannot open the image file")
					return
				}

				filename := ulid.Make().String() + path.Ext(f.Filename)

				img, err := assets.NewImage(filename, file, assets.WithWidthAndHeight(500, 500))

				if err != nil {
					log.Println("while uploading image")
				} else {
					images[index] = img
				}
			}(i)
		}

		wg.Wait()

		err = e.UpdateEvent(ctx.Context(), models.Event{ID: eventId, Images: images})
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "could not add images", err)
		}

		return pkg.Success(ctx, fiber.StatusOK)
	}
}
