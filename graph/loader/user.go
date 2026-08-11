package loader

import (
	"context"
	"net/http"
	"quizfreely/api/auth"
	"quizfreely/api/graph/model"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vikstrous/dataloadgen"
)

func (dr *dataReader) getUsers(ctx context.Context, userIDs []string) ([]*model.User, []error) {
	type dbUser struct {
		ID          *string `db:"id"`
		Username    *string `db:"username"`
		DisplayName *string `db:"display_name"`
	}
	var dbUsers []*dbUser

	err := pgxscan.Select(
		ctx,
		dr.db,
		&dbUsers,
		`SELECT u.id, u.username, u.display_name
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(id, og_order)
LEFT JOIN auth.users u ON u.id = input.id
ORDER BY input.og_order`,
		userIDs,
	)
	if err != nil {
		return nil, []error{err}
	}

	users := make([]*model.User, len(dbUsers))
	for i, du := range dbUsers {
		if du.ID == nil {
			users[i] = nil
		} else {
			users[i] = &model.User{
				ID:          du.ID,
				Username:    du.Username,
				DisplayName: du.DisplayName,
			}
		}
	}

	return users, nil
}

// GetUser returns single user by id efficiently
func GetUser(ctx context.Context, userID string) (*model.User, error) {
	loaders := For(ctx)
	return loaders.UserLoader.Load(ctx, userID)
}
// GetUsers returns many users by ids efficiently
func GetUsers(ctx context.Context, userIDs []string) ([]*model.User, error) {
	loaders := For(ctx)
	return loaders.UserLoader.LoadAll(ctx, userIDs)
}
