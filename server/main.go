package main

import (
	"context"
	"log"
	"net"

	pb "azuremachinelearning.com/scorer"
	"google.golang.org/grpc"
)

type scorerServer struct {
	pb.UnimplementedScorerServer
}

func (s *scorerServer) Score(ctx context.Context, in *pb.InferenceRequest) (*pb.InferenceResponse, error) {
	log.Printf("Received: %v", in.GetPrompt())
	return &pb.InferenceResponse{
		Result: in.GetPrompt() + " sunny",
	}, nil
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
