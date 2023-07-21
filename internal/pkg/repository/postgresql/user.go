package postgresql

import (
	"context"
	"database/sql"
	"go-redis/internal/pkg/db"
	"go-redis/internal/pkg/repository"
)

type UsersRepo struct {
	db db.DBops
}

func NewUsers(db db.DBops) *UsersRepo {
	return &UsersRepo{db: db}
}

func (r *UsersRepo) Add(ctx context.Context, name string) (int64, error) {
	var id int64
	err := r.db.ExecQueryRow(ctx,
		`INSERT INTO users(name) VALUES ($1) RETURNING id`,
		name).Scan(&id)
	return id, err
}

func (r *UsersRepo) GetById(ctx context.Context, id int64) (*repository.User, error) {
	var u repository.User
	err := r.db.Get(ctx, &u, "SELECT id,name FROM users WHERE id=$1", id)
	if err == sql.ErrNoRows {
		return nil, repository.ErrObjectNotFound
	}
	return &u, err
}

func (r *UsersRepo) Update(ctx context.Context, user *repository.User) (bool, error) {
	result, err := r.db.Exec(ctx,
		"UPDATE users SET name = $1 WHERE id = $2", user.Name, user.Id)
	return result.RowsAffected() > 0, err
}

func (r *UsersRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx,
		"DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}
