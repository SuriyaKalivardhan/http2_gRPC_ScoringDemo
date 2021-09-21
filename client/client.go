package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	pb "azuremachinelearning.com/scorer"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("[::1]:5001", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	testRPCtype := "Unary"
	if len(os.Args) < 2 {
		log.Printf("Not test RPC type provided, defaulting to Unary")
	} else {
		testRPCtype = os.Args[1]
		log.Printf("Testing: %s", testRPCtype)
	}

	client := pb.NewScorerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)

	switch testRPCtype {
	case "Unary":
		{
			testUnary(client, ctx)
			break
		}
	case "cStream":
		{
			testClientStreaming(client, ctx)
			break
		}
	case "sStream":
		{
			testServerStreaming(client, ctx)
			break
		}
	case "BiDi":
		{
			testBiDirectionStreaming(client, ctx)
			break
		}
	default:
		{
			log.Printf("No matching test found for %s, Supported values are Unary, cStream, sStream, BiDi", testRPCtype)
		}
	}
	defer cancel()
}

func testUnary(client pb.ScorerClient, ctx context.Context) {
	r, err := client.Score(ctx, &pb.InferenceRequest{Prompt: "Today is"})
	if err != nil {
		log.Fatalf("could not process: %v", err)
	}
	log.Printf("%s", r.GetResult())
}

func testClientStreaming(client pb.ScorerClient, ctx context.Context) {
	stream, err := client.StreamingRequestScore(ctx)
	if err != nil {
		log.Fatalf("Could not process client stream request: %v", err)
	}
	for i := 0; i < 11; i++ {
		prompt := fmt.Sprintf("%v", (i * i))
		log.Printf("Sending %v", prompt)
		stream.Send(&pb.InferenceRequest{
			Prompt: prompt,
		})
		time.Sleep(250 * time.Millisecond)
	}

	response, error := stream.CloseAndRecv()
	if error != nil {
		log.Fatalf("Did recived the response for client streamed request: %v", error)
	}
	log.Println(response)
}

func testServerStreaming(client pb.ScorerClient, ctx context.Context) {
	prompt := "Input size is "
	stream, err := client.StreamingResponseScore(ctx, &pb.InferenceRequest{
		Prompt: prompt,
	})

	if err != nil {
		log.Fatalf("Could not process server stream request: %v", err)
	}

	for {
		response, error := stream.Recv()
		if error == io.EOF {
			log.Println("Completed receiving all the response from Server")
			return
		} else if error != nil {
			log.Printf("Could not process server stream response: %v", error)
			return
		} else {
			log.Printf("Received response: %v", response)
		}
	}
}

func testBiDirectionStreaming(client pb.ScorerClient, ctx context.Context) {
	stream, error := client.BidirectionalScore(ctx)
	if error != nil {
		log.Printf("Error in Creating BiDirectional client %v", error)
		return
	}

	for i := 0; i < 10; i++ {
		err := stream.Send(&pb.InferenceRequest{
			Prompt: fmt.Sprintf("%v", (i * i)),
		})
		if err != nil {
			log.Printf("Error in Sending request in BiDirectional client %v", err)
		}

		if i%2 == 0 {
			response, err := stream.Recv()
			if err != nil {
				log.Printf("Error in receiving request in BiDirectional client %v", err)
			}
			log.Printf("Received %v", response.GetResult())
		}
		time.Sleep(250 * time.Millisecond)
	}
}
