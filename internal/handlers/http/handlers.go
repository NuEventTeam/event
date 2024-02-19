package http

import (
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	event_service "github.com/NuEventTeam/events/internal/services/event"
	"github.com/NuEventTeam/events/pkg"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"path/filepath"
	"time"
)

type Handler struct {
	eventSvc  *event_service.EventSvc
	JWTSecret string
}

func NewHttpHandler(
	svc *event_service.EventSvc,
	jwtSecret string,
) *Handler {

	return &Handler{
		eventSvc:  svc,
		JWTSecret: jwtSecret,
	}
}

func (h *Handler) SetUpRoutes(router *fiber.App) {
	apiV1 := router.Group("/api/v1")

	apiV1.Get("/categories", h.getCategories)

	apiV1.Post("/event/create", MustAuth(h.JWTSecret), h.createEvent)

	apiV1.Get("/event/show/:eventId", h.getEventByID)
}

type CreateEventRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Price       *float32 `json:"price"`
	Seats       *int64   `json:"seats"`
	MaxAge      *int64   `json:"max_age"`
	MinAge      *int64   `json:"min_age"`
	Address     string   `json:"address"`
	Longitude   float64  `json:"lg"`
	Latitude    float64  `json:"lt"`
	Date        string   `json:"date"`
	StartsAt    string   `json:"starts_at"`
	EndsAt      string   `json:"end_at"`
	Categories  []int64  `json:"categories"`
}

func (h *Handler) getEventByID(ctx *fiber.Ctx) error {
	eventID, err := ctx.ParamsInt("eventId")
	if err != nil {
		return pkg.Error(ctx, fiber.StatusInternalServerError, "not proper id", err)

	}

	event, err := h.eventSvc.GetEventByID(ctx.Context(), int64(eventID))
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
	}

	return pkg.Success(ctx, event)
}

func (h *Handler) createEvent(ctx *fiber.Ctx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		return err
	}

	userId := ctx.Locals("user_id").(int64)

	if len(form.Value["payload"]) == 0 {
		return pkg.Error(ctx, fiber.StatusBadRequest, "request does not contain payload field", fmt.Errorf("payload missing"))
	}
	payload := form.Value["payload"][0]

	var request CreateEventRequest

	err = sonic.ConfigFastest.Unmarshal([]byte(payload), &request)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
	}

	startTime, err := time.Parse(time.DateTime, request.StartsAt)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
	}

	if startTime.Before(time.Now().Add(time.Hour * 2)) {
		return pkg.Error(ctx, fiber.StatusBadRequest, "event start time mast be at least 2 hour before creations")
	}

	endTime, err := time.Parse(time.DateTime, request.EndsAt)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
	}

	if endTime.Before(startTime) {
		return pkg.Error(ctx, fiber.StatusBadRequest, "improper ending time")
	}

	var imgs []pkg.Content
	for _, f := range form.File["images"] {
		file, err := f.Open()
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "cannot open file", err)
		}
		imgs = append(imgs, pkg.Content{
			Fname:  "files",
			Ftype:  filepath.Ext(f.Filename),
			Reader: file,
			Size:   f.Size,
		})
	}

	var imgUrls []models.Image
	if len(imgs) > 0 {
		urls, err := pkg.UploadImages(userId, imgs...)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}
		images := []models.Image{}

		for _, i := range urls {
			images = append(images, models.Image{
				Url: i,
			})
		}
	}

	categories := []models.Category{}
	for _, c := range request.Categories {
		categories = append(categories, models.Category{
			ID: c,
		})
	}

	e := models.Event{
		Title:       request.Title,
		Description: request.Description,
		MaxAge:      request.MaxAge,
		MinAge:      request.MinAge,
		Images:      imgUrls,
		Categories:  categories,
		Locations: []models.Location{{
			Address:   request.Address,
			Longitude: request.Longitude,
			Latitude:  request.Latitude,
			Seats:     request.Seats,
			StartsAt:  startTime,
			EndsAt:    endTime,
		}},
		Managers: []models.Manager{{
			User: models.User{UserID: userId},
			Role: models.Role{
				Name:        pkg.AuthorTitle,
				Permissions: []int64{pkg.PermissionRead, pkg.PermissionVerify, pkg.PermissionVerify}},
		}},
		Attendees: nil,
	}
	eventID, err := h.eventSvc.CreateEvent(ctx.Context(), e)

	if err != nil {
		return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
	}

	return pkg.Success(ctx, fiber.Map{"event_id": eventID})
}

func (h *Handler) getCategories(ctx *fiber.Ctx) error {

	categories := []models.Category{}
	if ctx.QueryInt("all") == 1 {
		cats, err := h.eventSvc.GetCategoriesByID(ctx.Context(), nil)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}
		categories = cats
	}
	return pkg.Success(ctx, fiber.Map{"categories": categories})

}
