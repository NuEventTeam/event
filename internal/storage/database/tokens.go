package database

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/models"
	"time"
)

const TokensTable = "tokens"

func CreateToken(ctx context.Context, db DBTX, token models.Token) error {

	query := qb.Insert(TokensTable).
		Columns("token", "phone", "user_id", "token_type", "expires_at", "user_agent").
		Values(token.Token, token.Phone, token.UserId, token.Type, time.Now().Add(token.Duration), token.UserAgent)

	stmt, params, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)

	return err

}

func GetToken(ctx context.Context, db DBTX, token models.Token) (models.Token, error) {

	query := qb.Select("token,phone,user_id, token_type, user_agent").From(TokensTable).
		Where(sq.Gt{"expires_at": time.Now()}).
		Where(sq.Eq{"token_type": token.Type})

	if token.Token != "" {
		query = query.Where(sq.Eq{"token": token.Token})
	}

	if token.UserId != nil {
		query = query.Where(sq.Eq{"user_id": token.UserId})
	}

	if token.Phone != nil {
		query = query.Where(sq.Eq{"phone": token.Phone})
	}

	stmt, params, err := query.ToSql()
	if err != nil {
		return models.Token{}, err
	}

	var t models.Token
	err = db.QueryRow(ctx, stmt, params...).Scan(&t.Token, &t.Phone, &t.UserId, &t.Type, &t.UserAgent)
	if err != nil {
		return models.Token{}, err
	}

	return t, nil
}

func DeleteToken(ctx context.Context, db DBTX, token models.Token) error {

	query := qb.Delete(TokensTable).
		Where(sq.Eq{"token_type": token.Type})

	if token.Phone != nil {
		query = query.Where(sq.Eq{"phone": token.Phone})
	}

	if token.UserId != nil {
		query = query.Where(sq.Eq{"user_id": token.UserId})
	}

	if token.UserAgent != nil {
		query = query.Where(sq.Eq{"user_agent": token.UserAgent})
	}

	stmt, params, err := query.ToSql()

	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)
	return err
}
