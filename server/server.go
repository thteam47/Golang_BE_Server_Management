package server

import (
	"context"
	"fmt"
	"time"

	"github.com/thteam47/server_management/drive"
	models "github.com/thteam47/server_management/model"
	repo "github.com/thteam47/server_management/repository"
	"github.com/thteam47/server_management/serverpb"
)

type ServerImpl struct {
	serverpb.UnimplementedServerServiceServer
	userRepo   repo.UserRepository
	serverRepo repo.ServerRepository
	operaRepo  repo.OperationRepository
}

func NewServer(user repo.UserRepository, server repo.ServerRepository, opera repo.OperationRepository) *ServerImpl {
	return &ServerImpl{
		userRepo:   user,
		serverRepo: server,
		operaRepo:  opera,
	}
}
func (s *ServerImpl) Login(ctx context.Context, req *serverpb.LoginServer) (*serverpb.ResultLogin, error) {
	check, accessToken, role, err := s.userRepo.Login(req.GetUsername(), req.GetPassword())
	if check == false {
		return &serverpb.ResultLogin{
			Ok:          false,
			AccessToken: "",
			Role:        "",
		}, err
	}
	return &serverpb.ResultLogin{
		Ok:          true,
		AccessToken: accessToken,
		Role:        role,
	}, nil
}
func (s *ServerImpl) GetUser(ctx context.Context, req *serverpb.InfoUser) (*serverpb.User, error) {
	id, err := s.userRepo.GetIdUser(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println(id)
	user, err := s.userRepo.GetUser(id)
	if err != nil {
		return nil, err
	}
	return &serverpb.User{
		Username: user.Username,
		FullName: user.FullName,
		Password: user.Password,
		Email:    user.Email,
		IdUser:   user.ID.Hex(),
		Role:     user.Role,
		Action:   user.Action,
	}, nil
}
func (s *ServerImpl) GetListUser(ctx context.Context, req *serverpb.GetListUser) (*serverpb.ListUser, error) {
	listUser, err := s.userRepo.GetListUser()
	if err != nil {
		return nil, err
	}
	var dataResp []*serverpb.User
	for _, v := range listUser {
		elem := &serverpb.User{
			FullName: v.FullName,
			Email:    v.Email,
			Password: v.Password,
			Username: v.Username,
			Role:     v.Role,
			Action:   v.Action,
			IdUser:   v.ID.Hex(),
		}
		dataResp = append(dataResp, elem)
	}
	return &serverpb.ListUser{
		Data: dataResp,
	}, nil
}
func (s *ServerImpl) AddUser(ctx context.Context, req *serverpb.User) (*serverpb.User, error) {
	role := req.GetRole()
	var roleUser string
	if role == "" {
		roleUser = "staff"
	} else {
		roleUser = role
	}
	var actionList []string
	if roleUser == "admin" {
		actionList = append(actionList, "All Rights")
	} else if roleUser == "assistant" {
		actionList = []string{"Add Server", "Update Server", "Detail Status", "Export", "Connect", "Disconnect", "Delete Server", "Change Password"}
	} else {
		actionList = req.GetAction()
	}
	hashPass, _ := drive.HashPassword(req.GetPassword())
	id, err := s.userRepo.AddUser(&models.User{
		FullName: req.GetFullName(),
		Email:    req.GetEmail(),
		Password: hashPass,
		Username: req.GetUsername(),
		Role:     roleUser,
		Action:   actionList,
	})
	if err != nil {
		return nil, err
	}
	return &serverpb.User{
		IdUser:   id,
		FullName: req.GetFullName(),
		Email:    req.GetEmail(),
		Password: hashPass,
		Username: req.GetUsername(),
		Role:     roleUser,
		Action:   actionList,
	}, nil
}
func (s *ServerImpl) Connect(ctx context.Context, req *serverpb.LoginServer) (*serverpb.MessResponse, error) {
	err := s.operaRepo.Connect(req.GetUsername(), req.GetPassword())
	if err != nil {
		return &serverpb.MessResponse{
			Mess: "Username or password incorrect",
		}, err
	}
	return &serverpb.MessResponse{
		Mess: "Done",
	}, nil
}
func (s *ServerImpl) Index(ctx context.Context, req *serverpb.PaginationRequest) (*serverpb.ListServer, error) {
	var data []*models.Server
	var total int64
	data, total, err := s.serverRepo.Index(req.GetLimitPage(), req.GetNumberPage())
	fmt.Println(err)
	if err != nil {
		return nil, err
	}
	var dataResp []*serverpb.Server
	if total != 0 {
		for _, sv := range data {
			dataResp = append(dataResp, &serverpb.Server{
				IdServer:    sv.ID.Hex(),
				ServerName:  sv.ServerName,
				Username:    sv.Username,
				Port:        sv.Port,
				Ip:          sv.Ip,
				Password:    sv.Password,
				Description: sv.Description,
				Status:      sv.Status,
			})
		}
	}
	return &serverpb.ListServer{
		Data:        dataResp,
		TotalServer: total,
	}, nil
}
func (s *ServerImpl) Search(ctx context.Context, req *serverpb.SearchRequest) (*serverpb.ListServer, error) {
	var data []*models.Server
	var total int64
	data, total, err := s.serverRepo.Search(req.GetKeySearch(), req.GetFieldSearch(), req.GetLimitPage(), req.GetNumberPage())
	if err != nil {
		return nil, err
	}
	var dataResp []*serverpb.Server
	if total != 0 {
		for _, sv := range data {
			dataResp = append(dataResp, &serverpb.Server{
				IdServer:    sv.ID.Hex(),
				ServerName:  sv.ServerName,
				Username:    sv.Username,
				Port:        sv.Port,
				Ip:          sv.Ip,
				Password:    sv.Password,
				Description: sv.Description,
				Status:      sv.Status,
			})
		}
	}
	return &serverpb.ListServer{
		Data:        dataResp,
		TotalServer: total,
	}, nil
}
func (s *ServerImpl) CheckServerName(ctx context.Context, req *serverpb.CheckServerNameRequest) (*serverpb.CheckServerNameResponse, error) {
	check := s.serverRepo.CheckServerName(req.GetServerName())
	return &serverpb.CheckServerNameResponse{
		Check: check,
	}, nil
}
func (s *ServerImpl) AddServer(ctx context.Context, req *serverpb.Server) (*serverpb.Server, error) {
	server := &models.Server{
		ServerName:  req.GetServerName(),
		Username:    req.GetUsername(),
		Password:    req.GetPassword(),
		Ip:          req.GetIp(),
		Port:        req.GetPort(),
		Description: req.GetDescription(),
	}
	serverResp, err := s.serverRepo.AddServer(server)
	if err != nil {
		return nil, err
	}
	return &serverpb.Server{
		IdServer:    serverResp.ID.Hex(),
		ServerName:  serverResp.ServerName,
		Username:    serverResp.Username,
		Port:        serverResp.Port,
		Ip:          serverResp.Ip,
		Password:    serverResp.Password,
		Description: serverResp.Description,
	}, nil
}
func (s *ServerImpl) UpdateServer(ctx context.Context, req *serverpb.UpdateRequest) (*serverpb.Server, error) {
	server := &models.Server{
		ServerName:  req.GetInfoServer().GetServerName(),
		Username:    req.GetInfoServer().GetUsername(),
		Ip:          req.GetInfoServer().GetIp(),
		Port:        req.GetInfoServer().GetPort(),
		Description: req.GetInfoServer().GetDescription(),
	}
	serverResp, err := s.serverRepo.UpdateServer(req.GetIdServer(), server)
	if err != nil {
		return nil, err
	}
	return &serverpb.Server{
		IdServer:    serverResp.ID.Hex(),
		ServerName:  serverResp.ServerName,
		Username:    serverResp.Username,
		Port:        serverResp.Port,
		Ip:          serverResp.Ip,
		Password:    serverResp.Password,
		Description: serverResp.Description,
	}, nil
}
func (s *ServerImpl) DetailsServer(ctx context.Context, req *serverpb.DetailsServer) (*serverpb.DetailsServerResponse, error) {
	var data []*models.StatusDetail
	var status string
	
	status, data, err := s.serverRepo.DetailsServer(req.GetIdServer(), req.GetTimeIn(), req.GetTimeOut())
	if err != nil {
		return nil, err
	}
	var dataResp []*serverpb.StatusDetail
	for _, st := range data {
		dataResp = append(dataResp, &serverpb.StatusDetail{
			StatusDt: st.Status,
			Time:     st.Time.Format(time.RFC3339),
		})
	}
	return &serverpb.DetailsServerResponse{
		StatusServer: status,
		Status:       dataResp,
	}, nil
}
func (s *ServerImpl) DeleteServer(ctx context.Context, req *serverpb.DelServer) (*serverpb.DeleteServerResponse, error) {
	err := s.serverRepo.DeleteServer(req.GetIdServer())
	if err != nil {
		return &serverpb.DeleteServerResponse{
			Ok: false,
		}, err
	}
	return &serverpb.DeleteServerResponse{
		Ok: true,
	}, nil
}
func (s *ServerImpl) ChangePassword(ctx context.Context, req *serverpb.ChangePasswordRequest) (*serverpb.MessResponse, error) {
	err := s.serverRepo.ChangePassword(req.GetIdServer(), req.GetPassword())
	if err != nil {
		return &serverpb.MessResponse{
			Mess: "Error",
		}, err
	}
	return &serverpb.MessResponse{
		Mess: "Done",
	}, nil
}
func (s *ServerImpl) ChangeActionUser(ctx context.Context, req *serverpb.ChangeActionUser) (*serverpb.MessResponse, error) {
	err := s.userRepo.ChangeActionUser(req.GetIdUser(), req.GetRole(), req.GetAction())
	if err != nil {
		return nil, err
	}
	return &serverpb.MessResponse{
		Mess: "Done",
	}, nil
}
func (s *ServerImpl) CheckStatus(ctx context.Context, req *serverpb.CheckStatusRequest) (*serverpb.CheckStatusResponse, error) {
	return nil, nil
}
func (s *ServerImpl) Export(ctx context.Context, req *serverpb.ExportRequest) (*serverpb.ExportResponse, error) {
	url := s.operaRepo.Export(req.GetPage(), req.GetLimitPage(), req.GetNumberPage())
	return &serverpb.ExportResponse{
		Url: url,
	}, nil
}
func (s *ServerImpl) Logout(ctx context.Context, req *serverpb.Logout) (*serverpb.MessResponse, error) {
	check, err := s.userRepo.Logout(ctx)
	if check == false {
		return nil, err
	}
	return &serverpb.MessResponse{
		Mess: "Done",
	}, nil
}
func (s *ServerImpl) Disconnect(ctx context.Context, req *serverpb.Disconnect) (*serverpb.MessResponse, error) {
	err := s.operaRepo.Disconnect(req.GetIdServer())
	if err != nil {
		return &serverpb.MessResponse{
			Mess: "Id incorrect",
		}, err
	}
	return &serverpb.MessResponse{
		Mess: "Done",
	}, nil
}
func (s *ServerImpl) UpdateUser(ctx context.Context, req *serverpb.ChangeUser) (*serverpb.UserResponse, error) {
	user, err := s.userRepo.UpdateUser(req.GetIdUser(), &models.User{
		FullName: req.GetData().FullName,
		Email:    req.GetData().Email,
		Username: req.GetData().Username,
	})
	if err != nil {
		return nil, err
	}
	return &serverpb.UserResponse{
		IdUser: req.GetIdUser(),
		Data: &serverpb.User{
			FullName: user.FullName,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}
func (s *ServerImpl) ChangePassUser(ctx context.Context, req *serverpb.ChangePasswordUser) (*serverpb.MessResponse, error) {
	err := s.userRepo.ChangePassUser(req.GetIdUser(), req.GetPassword())
	if err != nil {
		return nil, err
	}
	return &serverpb.MessResponse{
		Mess: "Done",
	}, nil
}
func (s *ServerImpl) DeleteUser(ctx context.Context, req *serverpb.DeleteUser) (*serverpb.MessResponse, error) {

	err := s.userRepo.DeleteUser(req.GetIdUser())
	if err != nil {
		return nil, err
	}
	return &serverpb.MessResponse{
		Mess: "Done",
	}, nil
}
