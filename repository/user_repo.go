package repository

import (
	"context"

	models "github.com/thteam47/server_management/model"
)

type UserRepository interface {
	Login(username string, password string) (bool, string, string, error)
	Logout(ctx context.Context) (bool, error)
	GetListUser() ([]*models.User, error)
	GetUser(idUser string) (*models.User, error)
	AddUser(u *models.User) (string, error)
	ChangeActionUser(idUser string, role string, a []string) error
	UpdateUser(idUser string, u *models.User) (*models.User, error)
	ChangePassUser(idUser string, pass string) error
	DeleteUser(idUser string) error
	GetIdUser(ctx context.Context) (string, error)
}
