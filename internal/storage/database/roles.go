package database

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/models"
)

func CreateRole(ctx context.Context, db DBTX, role models.Role) (int64, error) {
	query := qb.Insert("event_roles").Columns("name", "event_id").Values(role.Name, role.EventID).Suffix("returning id")

	stmt, parms, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	var id int64

	err = db.QueryRow(ctx, stmt, parms...).Scan(&id)
	return id, err
}

func UpdateRole(ctx context.Context, db DBTX, roleID int64, name string) error {
	query := qb.Update("event_roles").Set("name", name)

	stmt, parms, err := query.ToSql()
	if err != nil {
		return err
	}

	res, err := db.Exec(ctx, stmt, parms...)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

func AddRolePermissions(ctx context.Context, db DBTX, role models.Role) error {

	query := qb.Insert("event_role_permissions").
		Columns("role_id", "permission_id")

	for _, i := range role.Permissions {
		query = query.Values(role.ID, i)
	}

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, args...)
	return err
}

func RemoveRolePermissions(ctx context.Context, db DBTX, roleId int64, permissionID ...int64) error {
	query := qb.Delete("event_role_permissions").
		Where(sq.Eq{"role_id": roleId}).
		Where(sq.Eq{"permission_id": permissionID[:]})

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, args...)
	return err
}

func GetRolePermissions(ctx context.Context, db DBTX, roleId int64) ([]int64, error) {

	query := qb.Select("permission_id").
		From("event_role_permissions").
		Where(sq.Eq{"role_id": roleId})

	stmt, params, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var permissions []int64

	rows, err := db.Query(ctx, stmt, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p int64
		err := rows.Scan(&p)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, p)

	}

	return permissions, nil
}
