package repository

import (
	"context"
	"errors"
)

var (
	ErrObjectNotFound = errors.New("object not found")
)

type UsersRepo interface {
	Add(ctx context.Context, name string) (int64, error)
	GetById(ctx context.Context, id int64) (*User, error)
	Update(ctx context.Context, user *User) (bool, error)
	Delete(ctx context.Context, id int64) error
}
