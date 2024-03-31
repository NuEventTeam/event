package comments

import (
	"context"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"time"
)

type AddCommentRequest struct {
	ParentId *int64 `json:"parentId"`
	Text     string `json:"text"`
	EventId  int64  `json:"eventId"`
	AuthorId int64  `json:"-"`
}

type Comment struct {
	CommentId int64     `json:"commentId"`
	ParentId  *int64    `json:"parentId"`
	Text      string    `json:"text"`
	Author    Author    `json:"author"`
	Children  []Comment `json:"children"`
	CreatedAt time.Time `json:"createdAt"`
}

type Author struct {
	ID           int64   `json:"id"`
	Username     string  `json:"username"`
	ProfileImage *string `json:"profileImage"`
}

func AddCommentHandler(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authorId := ctx.Locals("userId").(int64)

		var request AddCommentRequest
		if err := ctx.BodyParser(&request); err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "could not parse json", err)
		}
		request.AuthorId = authorId
		commentID, err := addComment(ctx.Context(), db.GetDb(), request)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "oops something went wrong", err)
		}

		author, err := getCommentAuthor(ctx.Context(), db.GetDb(), authorId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "oops something went wrong", err)
		}

		response := Comment{
			CommentId: commentID,
			ParentId:  request.ParentId,
			Text:      request.Text,
			Author:    author,
		}

		return pkg.Success(ctx, fiber.Map{"comment": response})
	}
}

func getCommentAuthor(ctx context.Context, db database.DBTX, userId int64) (Author, error) {
	query := `select id,username, profile_image from users where id = $1;`

	var author Author
	err := db.QueryRow(ctx, query, userId).Scan(&author.ID, &author.Username, &author.ProfileImage)
	if err != nil {
		return Author{}, err
	}
	if author.ProfileImage != nil {
		*author.ProfileImage = pkg.CDNBaseUrl + *author.ProfileImage
	}
	return author, nil

}

func addComment(ctx context.Context, db database.DBTX, request AddCommentRequest) (int64, error) {
	query := `insert into comments (parent_id, text, author_id, event_id) values ($1, $2, $3, $4) returning id;`

	args := []interface{}{request.ParentId, request.Text, request.AuthorId, request.EventId}

	var id int64

	err := db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}
