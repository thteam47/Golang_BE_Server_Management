package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/golang/glog"
	"github.com/soheilhy/cmux"
	"github.com/thteam47/server_management/client"
	servergrpc "github.com/thteam47/server_management/serverGrpc"
	"golang.org/x/sync/errgroup"
)

func clientGolang(lis net.Listener) error {
	flag.Parse()
	defer glog.Flush()
	return client.Run(lis)
}

func serverGolang(lis net.Listener) error {
	flag.Parse()
	defer glog.Flush()
	return servergrpc.Run(lis)
}
func main() {
	fmt.Println("running")

	for {
		lis, err := net.Listen("tcp", ":9090")
		if err != nil {
			log.Fatalf("err while create listen %v", err)
		}
		m := cmux.New(lis)
		grpcListener := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
		httpListener := m.Match(cmux.HTTP1Fast())
		g := new(errgroup.Group)
		g.Go(func() error { return serverGolang(grpcListener) })
		g.Go(func() error { return clientGolang(httpListener) })
		g.Go(func() error { return m.Serve() })
		log.Println("run server:", g.Wait())
	}
}
