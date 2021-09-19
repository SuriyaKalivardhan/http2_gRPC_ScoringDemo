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

func (s *scorerServer) BidirectionalScore(stream pb.Scorer_BidirectionalScoreServer) error {
	log.Println("Starting the bidirectional request processing")
	result := []string{"BATCH START "}
	for i := 0; i < 10; i++ {
		request, error := stream.Recv()
		if error != nil {
			log.Printf("Could not process bidirection request %v", error)
			return error
		} else {
			result = append(result, request.GetPrompt())
		}

		if i%2 == 0 {
			batchResult := strings.Join(result, "__") + " BATCH END"
			log.Printf("Sending current stream response %s", batchResult)
			stream.Send(&pb.InferenceResponse{
				Result: batchResult,
			})
			result = []string{"BATCH START "}
		}
		time.Sleep(125 * time.Millisecond)
	}
	log.Println("Ending the bidirectional request processing")
	return nil
}
