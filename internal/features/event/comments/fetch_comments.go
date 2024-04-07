package comments

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
)

var qb = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type FetchCommentRequest struct {
	LastParentId int64 `json:"lastParentId"`
	EventId      int64 `json:"eventId"`
}

func FetchCommentHandler(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {

		var request FetchCommentRequest

		if err := ctx.BodyParser(&request); err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "cannot parse json", err)
		}

		parentComments, parentIds, err := getParentComments(ctx.Context(), db.GetDb(), request)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "oops unexpected error", err)
		}

		childComments, err := getChildComments(ctx.Context(), db.GetDb(), parentIds)

		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "oops unexpected error", err)
		}

		for i := 0; i < len(parentComments); i++ {
			if val, ok := childComments[parentComments[i].CommentId]; ok {
				parentComments[i].Children = val
			}
		}
		return pkg.Success(ctx, fiber.Map{"comments": parentComments})
	}
}

func getParentComments(ctx context.Context, db database.DBTX, param FetchCommentRequest) ([]Comment, []int64, error) {
	//TODO switch to sq

	query := qb.Select("comments.id", "comments.text", "comments.parent_id",
		"users.id", "users.profile_image", "users.username", "comments.created_at").
		From("comments").
		InnerJoin("users on users.id = comments.author_id")

	if param.LastParentId != 0 {
		query = query.Where(sq.Lt{"comments.id": param.LastParentId})
	}

	query = query.Where(sq.Eq{"comments.event_id": param.EventId}).
		OrderBy("comment.id desc", "comments.created_at desc")

	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, nil, err
	}

	rows, err := db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()
	var comments []Comment
	var parentIds []int64
	for rows.Next() {
		var c Comment
		err := rows.Scan(&c.CommentId, &c.Text, &c.ParentId, &c.Author.ID, &c.Author.ProfileImage, &c.Author.Username, &c.CreatedAt)
		if err != nil {
			return nil, nil, err
		}

		if c.Author.ProfileImage != nil {
			*c.Author.ProfileImage = pkg.CDNBaseUrl + "/get/" + *c.Author.ProfileImage
		}

		comments = append(comments, c)
		parentIds = append(parentIds, c.CommentId)
	}

	return comments, parentIds, err
}

func getChildComments(ctx context.Context, db database.DBTX, parentIds []int64) (map[int64][]Comment, error) {
	query := qb.Select("comments.id", "comments.text", "comments.parent_id",
		"users.id", "users.profile_image", "users.username", "comments.created_at").
		From("comments").
		InnerJoin("users on users.id = comments.author_id").
		Where(sq.Eq{"parent_id": parentIds}).OrderBy("parent_id desc", "comments.created_at desc")

	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	m := map[int64][]Comment{}

	for rows.Next() {
		var c Comment
		err := rows.Scan(&c.CommentId, &c.Text, &c.ParentId, &c.Author.ID, &c.Author.ProfileImage, &c.Author.Username, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		if c.Author.ProfileImage != nil {
			*c.Author.ProfileImage = pkg.CDNBaseUrl + *c.Author.ProfileImage
		}
		if _, ok := m[*c.ParentId]; !ok {
			m[*c.ParentId] = []Comment{}
		}
		m[*c.ParentId] = append(m[*c.ParentId], c)
	}
	return m, nil
}
