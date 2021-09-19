package main

import (
	"log"
	"net"

	pb "azuremachinelearning.com/scorer"
	"google.golang.org/grpc"
)

type scorerServer struct {
	pb.UnimplementedScorerServer
}

func main() {
	listener, err := net.Listen("tcp", "localhost:55555")
	if err != nil {
		log.Printf("Exception occured %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterScorerServer(grpcServer, &scorerServer{})
	grpcServer.Serve(listener)
}
