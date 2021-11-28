package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"path"
	"strings"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	gw "github.com/thteam47/server_management/serverpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	grpcServerEndpoint = flag.String("grpc-server-endpoint", ":9090", "gRPC server endpoint")
)

func Run(lis net.Listener) error {
	flag.Parse()
	log.Printf("dial server %s", *grpcServerEndpoint)
	transportOption := grpc.WithInsecure()

	// tlsCredentials, err := loadTLSCredentials()
	// if err != nil {
	// 	log.Fatal("cannot load TLS credentials: ", err)
	// }

	// /transportOption := grpc.WithTransportCredentials(tlsCredentials)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	dialOpts := []grpc.DialOption{transportOption}
	gwmux := runtime.NewServeMux()
	err := gw.RegisterServerServiceHandlerFromEndpoint(ctx, gwmux, *grpcServerEndpoint, dialOpts)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle("/", gwmux)
	mux.HandleFunc("/swagger/", serveSwaggerFile)
	serveSwaggerUI(mux)
	s := &http.Server{Handler: allowCORS(mux)}
	return s.Serve(lis)
}
func serveSwaggerUI(mux *http.ServeMux) {
	fs := http.FileServer(http.Dir("./swaggerui/"))
	prefix := "/swaggerui/"
	mux.Handle(prefix, http.StripPrefix(prefix, fs))
}
func serveSwaggerFile(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "swagger.json") {
		fmt.Printf("Not Found: %s\r\n", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	p := strings.TrimPrefix(r.URL.Path, "./swaggerui/")
	p = path.Join("../protos", p)
	fmt.Printf("Serving swagger-file: %s\r\n", p)
	http.ServeFile(w, r, p)
}
func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "Authorization"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	glog.Infof("preflight request for %s", r.URL.Path)
	return
}
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}
func loadTLSCredentials() (credentials.TransportCredentials, error) {

	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := ioutil.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair("cert/client-cert.pem", "cert/client-key.pem")
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}
