package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	pb "azuremachinelearning.com/scorer"
	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:55555")
	if err != nil {
		log.Printf("Exception occured %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterScorerServer(grpcServer, &scorerServer{})
	grpcServer.Serve(listener)
}

type scorerServer struct {
	pb.UnimplementedScorerServer
}

func (s *scorerServer) Score(ctx context.Context, request *pb.InferenceRequest) (*pb.InferenceResponse, error) {
	log.Printf("Received: %v", request.GetPrompt())
	return &pb.InferenceResponse{
		Result: request.GetPrompt() + " sunny",
	}, nil
}

func (s *scorerServer) StreamingRequestScore(stream pb.Scorer_StreamingRequestScoreServer) error {
	result := []string{"START "}
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			finalResult := strings.Join(result, "__") + " END"
			log.Printf("End of streaming request, will return the response %s", finalResult)
			return stream.SendAndClose(&pb.InferenceResponse{
				Result: finalResult,
			})
		}
		if err != nil {
			return nil
		}
		if len(result) == 1 {
			log.Println("First response received from client")
		}
		result = append(result, request.GetPrompt())
	}
}

func (s *scorerServer) StreamingResponseScore(request *pb.InferenceRequest, stream pb.Scorer_StreamingResponseScoreServer) error {
	prompt := request.GetPrompt()
	log.Println("Sending first response for the Server Streaming request")
	for i := 0; i < 10; i++ {
		result := fmt.Sprintf("%s %v", prompt, i)
		error := stream.Send(&pb.InferenceResponse{
			Result: result,
		})
		if error != nil {
			log.Printf("Error in processing Server streaming request %v", error)
		}
		time.Sleep(250 * time.Millisecond)
	}
	log.Println("Sent all the request to client")
	return nil
}
