package repository

import (
	models "github.com/thteam47/server_management/model"
)

type ServerRepository interface {
	Index(limitpage int64, numberpage int64) ([]*models.Server, int64, error)
	AddServer(sv *models.Server) (*models.Server, error)
	UpdateServer(idServer string, sv *models.Server) (*models.Server, error)
	DetailsServer(idServer string, timeIn string, timeOut string) (string, []*models.StatusDetail, error)
	DeleteServer(idServer string) error
	ChangePassword(idServer string, pass string) error
	CheckServerName(servername string) bool
	Search(key string,field string, limitpage int64, numberpage int64) ([]*models.Server, int64, error)
	UpdateStatus()
}
