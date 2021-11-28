package repository

import (
	models "github.com/thteam47/server_management/model"
)

type JwtRepository interface {
	Generate(user *models.User) (string, error)
	Verify(accessToken string) (*models.Claims, error)
}
