package user

import (
	"fmt"
	"github.com/NuEventTeam/events/internal/features/assets"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/oklog/ulid/v2"
	"log"
	"path"
	"time"
)

func (u User) UpdateUserHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		form, err := ctx.MultipartForm()
		if err != nil {
			return err
		}

		userId := ctx.Locals("userId").(int64)

		var request MobileUser
		if len(form.Value["payload"]) == 0 {
			return pkg.Error(ctx, fiber.StatusBadRequest, "payload is missing", fmt.Errorf("payload is missing"))
		}

		err = sonic.ConfigFastest.Unmarshal([]byte(form.Value["payload"][0]), &request)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "cannot read payload")
		}

		exists, err := u.CheckUsername(ctx.Context(), request.Username)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}
		if exists {
			return pkg.Error(ctx, fiber.StatusBadRequest, "username exists")
		}

		exists, err = u.CheckUserID(ctx.Context(), userId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}
		if !exists {
			return pkg.Error(ctx, fiber.StatusBadRequest, "user does not exists")
		}

		var image assets.Image

		for _, f := range form.File["images"] {

			file, err := f.Open()
			if err != nil {
				log.Println("cannot open the image file")
				break
			}

			filename := pkg.UserNamespace + "/" + fmt.Sprintf("%d", userId) + "/" + ulid.Make().String() + path.Ext(f.Filename)

			img, err := assets.NewImage(filename, file, assets.WithWidthAndHeight(500, 500))
			if err != nil {
				log.Println("while uploading image")
				break
			}
			image = img

		}

		request.UserID = userId
		request.ProfileImage = image.Filename

		birthDate, err := time.Parse(time.DateOnly, request.BirthDate)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)

		}

		var categoryIds []models.Category
		for _, id := range request.Preferences {
			categoryIds = append(categoryIds, models.Category{ID: id})
		}

		if pkg.MinCategories > len(categoryIds) {
			return pkg.Error(ctx, fiber.StatusBadRequest, "must include 3 cats", fmt.Errorf("minimum of 3 cats"))

		}
		err = u.CreateUser(ctx.Context(), models.User{
			ID:           userId,
			UserID:       userId,
			Phone:        request.Phone,
			Username:     request.Username,
			Firstname:    request.Firstname,
			Lastname:     request.Lastname,
			Image:        image,
			ProfileImage: request.ProfileImage,
			BirthDate:    birthDate,
			Preferences:  categoryIds,
		})
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}
		user, err := database.GetUser(ctx.Context(), u.db.GetDb(), database.GetUserArgss{UserID: &userId})
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}
		if user.ProfileImage != nil {
			profileImgUrl := fmt.Sprint(pkg.CDNBaseUrl, *user.ProfileImage)
			user.ProfileImage = &profileImgUrl
		}
		return pkg.Success(ctx, fiber.Map{"username": request.Username, "user": user})
	}
}
