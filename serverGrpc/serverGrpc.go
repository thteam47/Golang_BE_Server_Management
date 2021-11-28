package servergrpc

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/spf13/viper"
	"github.com/thteam47/server_management/drive"
	repoimpl "github.com/thteam47/server_management/repository/repoImpl"
	"github.com/thteam47/server_management/server"
	"github.com/thteam47/server_management/serverpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

func accessibleRoles() map[string][]string {
	const serverServicePath = "/server.management.ServerService/"

	return map[string][]string{
		serverServicePath + "addServer":        {"admin", "assistant", "Add Server"},
		serverServicePath + "detailsServer":    {"admin", "assistant", "Detail Status"},
		serverServicePath + "updateServer":     {"admin", "assistant", "Update Server"},
		serverServicePath + "deleteServer":     {"admin", "assistant", "Delete Server"},
		serverServicePath + "changePassword":   {"admin", "assistant", "Change Password"},
		serverServicePath + "changePassUser":   {"admin", "assistant", "staff"},
		serverServicePath + "connect":          {"admin", "assistant", "Connect"},
		serverServicePath + "disconnect":       {"admin", "assistant", "Disconnect"},
		serverServicePath + "export":           {"admin", "assistant", "Export"},
		serverServicePath + "getUser":          {"admin", "assistant", "staff"},
		serverServicePath + "updateUser":       {"admin", "assistant", "staff"},
		serverServicePath + "checkServerName":  {"admin", "assistant"},
		serverServicePath + "search":           {"admin", "assistant"},
		serverServicePath + "deleteUser":       {"admin"},
		serverServicePath + "addUser":          {"admin"},
		serverServicePath + "changeActionUser": {"admin"},
		serverServicePath + "getListUser":      {"admin"},
	}
}

const (
	serverCertFile   = "cert/server-cert.pem"
	serverKeyFile    = "cert/server-key.pem"
	clientCACertFile = "cert/ca-cert.pem"
)
const (
	tokenDuration = 30 * time.Minute
)

func Run(lis net.Listener) error {
	vi := viper.New()
	vi.SetConfigFile("config.yaml")
	vi.ReadInConfig()
	flag.Parse()
	db := drive.ConnectMongo(vi.GetString("dburl"), vi.GetString("dbname"))
	redis := drive.ConnectRedis(vi.GetString("dbredis"))
	elas := drive.ConnectElasticsearch(vi.GetString("dbelasticurl"))
	jwtManager := repoimpl.NewJwtRepo(vi.GetString("secretKey"), vi.GetDuration("tokenDuration"))
	userRepo := repoimpl.NewUserRepo(db.DB, redis.MyRediscache, jwtManager)
	serverRepo := repoimpl.NewServerRepo(db.DB, redis.MyRediscache, elas.Elas)
	operaRepo := repoimpl.NewOperationRepo(db.DB, redis.MyRediscache, elas.Elas)
	interceptor := repoimpl.NewAuthInterceptor(jwtManager, accessibleRoles())

	//go serverRepo.UpdateStatus()
	//go operaRepo.SendMail()

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		//grpc.StreamInterceptor(interceptor.Stream()),
	}
	// tlsCredentials, err := loadTLSCredentials()
	// if err != nil {
	// 	return fmt.Errorf("cannot load TLS credentials: %w", err)
	// }
	// serverOptions = append(serverOptions, grpc.Creds(tlsCredentials))
	s := grpc.NewServer(serverOptions...)
	serverpb.RegisterServerServiceServer(s, server.NewServer(userRepo, serverRepo, operaRepo))
	reflection.Register(s)
	return s.Serve(lis)
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed client's certificate
	pemClientCA, err := ioutil.ReadFile(clientCACertFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}
