package main

import (
	"flag"
	"log"
	"net"

	"github.com/artistomin/proxy/internal/app/activity/infrastructure/grpc"
	"github.com/artistomin/proxy/internal/app/activity/infrastructure/postgresql"
	"github.com/artistomin/proxy/internal/app/activity/usecase/request"
	"github.com/artistomin/proxy/internal/app/activity/usecase/response"
)

var (
	dsn  = flag.String("dsn", "postgres://root:root@localhost:5432/activity?sslmode=disable", "Database URI")
	port = flag.String("port", ":3000", "grpc server port")
)

func main() {
	flag.Parse()

	db, err := postgresql.New(*dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	reqRepo := postgresql.NewRequestRepo(db)
	resRepo := postgresql.NewResponseRepo(db)

	reqSvc := request.NewService(reqRepo)
	resSvc := response.NewService(resRepo)

	grpcServer := grpc.NewServer(reqSvc, resSvc)

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	log.Printf("Listening grpc on %s\n", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve grpc: %s", err)
	}
}
