package main

import (
	"context"
	"log"
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

	client := pb.NewScorerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := client.Score(ctx, &pb.InferenceRequest{Prompt: "Today is"})
	if err != nil {
		log.Fatalf("could not process: %v", err)
	}
	log.Printf("%s", r.GetResult())
}
