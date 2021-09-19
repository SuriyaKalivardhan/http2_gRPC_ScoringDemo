package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "azuremachinelearning.com/scorer"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:55555", grpc.WithInsecure(), grpc.WithBlock())
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	switch testRPCtype {
	case "Unary":
		{
			testUnary(client, ctx)
			break
		}
	case "cStream":
		{
			testUnary(client, ctx)
			break
		}
	case "sStream":
		{
			testUnary(client, ctx)
			break
		}
	case "BiDi":
		{
			testUnary(client, ctx)
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
