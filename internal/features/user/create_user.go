package user

import (
	"context"
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
	"strconv"
	"time"
)

type MobileUser struct {
	UserID       int64   `json:"userID,omitempty"`
	Username     string  `json:"username,omitempty"`
	ProfileImage *string `json:"profileImage,omitempty"`
	BirthDate    string  `json:"birthDate,omitempty"`
	Phone        string  `json:"phone,omitempty"`
	Firstname    string  `json:"firstname,omitempty"`
	Lastname     *string `json:"lastname,omitempty"`
	Preferences  []int64 `json:"preferences,omitempty"`
}

func (u User) CreateMobileUserHandler() fiber.Handler {
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

		var image *assets.Image

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
		request.ProfileImage = &image.Filename

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

		return pkg.Success(ctx, fiber.Map{"username": request.Username})
	}
}

func (u User) CreateUserHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		form, err := ctx.MultipartForm()
		if err != nil {
			return err
		}

		userId := ctx.Locals("userId").(int64)

		exists, err := u.CheckUsername(ctx.Context(), form.Value["username"][0])
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

		var categoryIds []models.Category

		for _, i := range form.Value["category_ids"] {
			id, _ := strconv.ParseInt(i, 10, 64)
			categoryIds = append(categoryIds, models.Category{ID: id})
		}

		if pkg.MinCategories > len(categoryIds) {
			return pkg.Error(ctx, fiber.StatusBadRequest, "must include 3 cats", fmt.Errorf("minimum of 3 cats"))

		}
		_, err = time.Parse(time.DateOnly, form.Value["birthdate"][0])
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}

		var image *assets.Image

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
			break

		}
		log.Println(image)
		birthDate, err := time.Parse(time.DateOnly, form.Value["birthdate"][0])
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)

		}
		err = u.CreateUser(ctx.Context(), models.User{
			ID:           userId,
			UserID:       userId,
			Username:     form.Value["username"][0],
			ProfileImage: &image.Filename,
			BirthDate:    birthDate,
			Phone:        form.Value["phone"][0],
			Firstname:    form.Value["firstname"][0],
			Lastname:     &form.Value["lastname"][0],
			Preferences:  categoryIds,
			Image:        image,
		})
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}

		return pkg.Success(ctx, fiber.Map{"username": form.Value["username"][0]})

	}
}

func (u User) CreateUser(ctx context.Context, user models.User) error {
	tx, err := u.db.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	err = database.CreateUser(ctx, tx, user)
	if err != nil {
		return err
	}

	u.assets.Upload(ctx, user.Image)

	err = database.AddUserPreference(ctx, tx, user.ID, user.Preferences...)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (e *User) CheckUsername(ctx context.Context, username string) (bool, error) {
	exists, err := database.CheckUsername(ctx, e.db.GetDb(), username)
	return exists, err
}
func (e *User) CheckUserID(ctx context.Context, userID int64) (bool, error) {
	exists, err := database.CheckUserExists(ctx, e.db.GetDb(), userID)
	return exists, err
}
