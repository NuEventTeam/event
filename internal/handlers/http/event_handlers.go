package http

import (
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/services/cdn"
	event_service "github.com/NuEventTeam/events/internal/services/event"
	"github.com/NuEventTeam/events/internal/services/users"
	"github.com/NuEventTeam/events/pkg"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/oklog/ulid/v2"
	"path"
	"strconv"
	"time"
)

type Handler struct {
	userSvc   *users.UserService
	eventSvc  *event_service.EventSvc
	cdnSvc    *cdn.CdnSvc
	JWTSecret string
}

func NewHttpHandler(
	eventSvc *event_service.EventSvc,
	cdn *cdn.CdnSvc,
	userSvc *users.UserService,
	jwtSecret string,
) *Handler {

	return &Handler{
		eventSvc:  eventSvc,
		cdnSvc:    cdn,
		userSvc:   userSvc,
		JWTSecret: jwtSecret,
	}
}

func (h *Handler) SetUpEventRoutes(router *fiber.App) {
	apiV1 := router.Group("/api/v1")

	apiV1.Get("/categories", h.getCategories)

	apiV1.Post("/event/create", MustAuth(h.JWTSecret), h.createEvent)

	apiV1.Get("/event/show/:eventId", h.getEventByID)

	apiV1.Put("/event/posts/:eventId",
		MustAuth(h.JWTSecret),
		h.HasPermission(pkg.PermissionUpdate),
		h.updateEvent,
	)

	apiV1.Put("/event/image/:eventId",
		MustAuth(h.JWTSecret),
		h.HasPermission(pkg.PermissionUpdate),
		h.addImage,
	)

	apiV1.Put("/event/fellowship/follow/:eventId",
		MustAuth(h.JWTSecret),
		h.followEvent)

	apiV1.Put("/event/fellowship/unfollow/:eventId",
		MustAuth(h.JWTSecret),
		h.unfollowEvent)
}
func (h *Handler) followEvent(ctx *fiber.Ctx) error {
	userId := ctx.Locals("userId").(int64)
	eventId, err := strconv.ParseInt(ctx.Params("eventId"), 10, 64)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, "invalid event id", err)
	}

	err = h.eventSvc.CheckEventStatus(ctx.Context(), eventId)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
	}
	err = h.eventSvc.AddFollower(ctx.Context(), eventId, userId)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, "something went wrong", err)
	}
	return pkg.Success(ctx, nil)
}

func (h *Handler) unfollowEvent(ctx *fiber.Ctx) error {
	userId := ctx.Locals("userId").(int64)
	eventId, err := strconv.ParseInt(ctx.Params("eventId"), 10, 64)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, "invalid event id", err)
	}

	err = h.eventSvc.RemoveFollower(ctx.Context(), eventId, userId)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, "something went wrong", err)
	}
	return pkg.Success(ctx, nil)
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

	userId := ctx.Locals("userId").(int64)

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

	var uploadContent []cdn.Content

	for _, f := range form.File["images"] {
		file, err := f.Open()
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "cannot open file", err)
		}
		filename := ulid.Make().String() + path.Ext(f.Filename)
		uploadContent = append(uploadContent, cdn.Content{
			FieldName: "files",
			Filename:  filename,
			Payload:   file,
			Size:      f.Size,
		})
	}

	e := models.Event{
		Title:       &request.Title,
		Description: &request.Description,
		MaxAge:      request.MaxAge,
		MinAge:      request.MinAge,
		Images:      make([]models.Image, len(uploadContent)),
		CategoryIds: request.Categories,
		Locations: []models.Location{{
			Address:   &request.Address,
			Longitude: &request.Longitude,
			Latitude:  &request.Latitude,
			Seats:     request.Seats,
			StartsAt:  &startTime,
			EndsAt:    &endTime,
		}},
		Managers: []models.Manager{{
			User: models.User{UserID: userId},
			Role: models.Role{
				Name:        pkg.AuthorTitle,
				Permissions: []int64{pkg.PermissionRead, pkg.PermissionVerify, pkg.PermissionUpdate}},
		}},
		Attendees:    nil,
		ImageContent: uploadContent,
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

type RemoveImageRequest struct {
	Images []int64 `json:"imageIds"`
}

func (h *Handler) removeEventImage(ctx *fiber.Ctx) error {
	var request RemoveImageRequest
	if err := ctx.BodyParser(&request); err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, "invalid json", err)
	}

	eventId, err := strconv.ParseInt(ctx.Params("eventId"), 10, 64)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, "invalid eventID", err)
	}

	urls, err := h.eventSvc.RemoveImage(ctx.Context(), eventId, request.Images...)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusInternalServerError, "could not remove image from db ", err)
	}

	h.cdnSvc.Delete(urls...)

	return pkg.Success(ctx, nil)
}

type UpdateEventRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Location    *struct {
		ID        int64    `json:"locationId"`
		Address   *string  `json:"address"`
		Longitude *float64 `json:"longitude"`
		Latitude  *float64 `json:"latitude"`
		StartsAt  *string  `json:"startsAt"`
		Seats     *int64   `json:"seats"`
		EndsAt    *string  `json:"endsAt"`
	} `json:"location"`
	Categories     []int64 `json:"categories"`
	AgeMax         *int64  `json:"ageMax"`
	AgeMin         *int64  `json:"ageMin"`
	Status         *int    `json:"status"`
	RemoveImageIds []int64 `json:"removeImages"`
}

func (h *Handler) updateEvent(ctx *fiber.Ctx) error {
	var request UpdateEventRequest

	if err := ctx.BodyParser(&request); err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, "invalid json", err)
	}

	eventId, err := ctx.ParamsInt("eventId")
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, "invalid event id", err)
	}

	event := models.Event{
		ID:              int64(eventId),
		Title:           request.Title,
		Description:     request.Description,
		MaxAge:          request.AgeMax,
		MinAge:          request.AgeMin,
		CategoryIds:     request.Categories,
		Status:          request.Status,
		RemoveImagesIds: request.RemoveImageIds,
	}

	if request.Location != nil {
		l := request.Location
		var startsAt, endsAt *time.Time
		if l.StartsAt != nil {
			t, err := time.Parse(time.DateTime, *l.StartsAt)
			if err != nil {
				return pkg.Error(ctx, fiber.StatusBadRequest, "invalid start date", err)
			}
			startsAt = &t
		}
		if l.EndsAt != nil {
			t, err := time.Parse(time.DateTime, *l.EndsAt)
			if err != nil {
				return pkg.Error(ctx, fiber.StatusBadRequest, "invalid start date", err)
			}
			endsAt = &t
		}
		event.Locations = append(event.Locations, models.Location{
			ID:        l.ID,
			EventID:   int64(eventId),
			Address:   l.Address,
			Longitude: l.Longitude,
			Latitude:  l.Latitude,
			StartsAt:  startsAt,
			EndsAt:    endsAt,
			Seats:     l.Seats,
		})
	}

	err = h.eventSvc.UpdateEvent(ctx.Context(), event)
	if err != nil {
		return err
	}
	return pkg.Success(ctx, nil)
}

func (h *Handler) addImage(ctx *fiber.Ctx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		return err
	}

	eventId, err := strconv.ParseInt(ctx.Params("eventId"), 10, 64)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, "invalid eventID", err)
	}
	var (
		uploadContent []cdn.Content
		imgUrls       []models.Image
	)
	for _, f := range form.File["images"] {
		file, err := f.Open()
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "cannot open file", err)
		}
		filename := ulid.Make().String() + path.Ext(f.Filename)
		uploadContent = append(uploadContent, cdn.Content{
			FieldName: "files",
			Filename:  filename,
			Payload:   file,
			Size:      f.Size,
		})
		imgUrls = append(imgUrls, models.Image{
			Url: fmt.Sprint(pkg.EventNamespace, eventId, "/", filename),
		})
	}

	err = h.eventSvc.UpdateEvent(ctx.Context(), models.Event{ID: eventId, Images: imgUrls, ImageContent: uploadContent})
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, "could not add images", err)
	}

	return pkg.Success(ctx, fiber.StatusOK)
}
