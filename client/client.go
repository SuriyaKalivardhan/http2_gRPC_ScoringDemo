package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	pb "azuremachinelearning.com/scorer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type tokenAuth struct {
	token string
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (tokenAuth) RequireTransportSecurity() bool {
	return true
}

func main() {
	//conn, err := grpc.Dial(":5001", grpc.WithInsecure(), grpc.WithBlock())
	//conn, err := grpc.Dial("https://ep-suriyak-onebox-2.eastus.inference.ml.azure.com", grpc.WithInsecure(), grpc.WithBlock())
	//conn, err := grpc.Dial("suriyakvm.westus2.cloudapp.azure.com:5001", grpc.WithInsecure(), grpc.WithBlock())

	tlsConfig := &tls.Config{}
	tlsConfig.InsecureSkipVerify = true

	token := "eyJ0eXAiOi..."
	conn, err := grpc.Dial("ep-suriyak-onebox-2.eastus.inference.ml.azure.com:443",
		grpc.WithTransportCredentials( //credentials.NewClientTLSFromCert(insecure.CertPool, "")),
			credentials.NewTLS(tlsConfig)),
		grpc.WithPerRPCCredentials(tokenAuth{
			token: token,
		}))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	testRPCtype := "Unary"
	if len(os.Args) < 2 {
		log.Printf("Not test RPC type provided, defaulting to Unary")
	} else {
		testRPCtype = os.Args[1]
	}

	client := pb.NewScorerClient(conn)
	for {
		log.Printf("Testing: %s", testRPCtype)
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
		case "All":
			{
				for i := 1; i <= 1; i++ {
					go testAll(client, ctx)
				}
				break
			}
		default:
			{
				log.Printf("No matching test found for %s, Supported values are Unary, cStream, sStream, BiDi, All", testRPCtype)
			}
		}

		reader := bufio.NewReader(os.Stdin)
		log.Print("Enter next test type  Unary, cStream, sStream, BiDi, All, Exit: ")
		text, _ := reader.ReadString('\n')
		testRPCtype = strings.Trim(text, "\r\n")
		if testRPCtype == "Exit" || testRPCtype == "exit" {
			log.Println("Exiting from program, closing the connection")
			cancel()
			conn.Close()
			return
		}

	}
}

func testUnary(client pb.ScorerClient, ctx context.Context) {
	r, err := client.Score(ctx, &pb.InferenceRequest{Prompt: "Today is"})
	if err != nil {
		log.Fatalf("could not process: %v", err)
	}
	log.Printf("Unary result %s", r.GetResult())
}

func testClientStreaming(client pb.ScorerClient, ctx context.Context) {
	stream, err := client.StreamingRequestScore(ctx)
	if err != nil {
		log.Fatalf("Could not process client stream request: %v", err)
	}
	for i := 0; i < 11; i++ {
		prompt := fmt.Sprintf("%v", (i * i))
		log.Printf("cStream Sending %v", prompt)
		stream.Send(&pb.InferenceRequest{
			Prompt: prompt,
		})
		time.Sleep(250 * time.Millisecond)
	}

	response, error := stream.CloseAndRecv()
	if error != nil {
		log.Fatalf("Did recived the response for client streamed request: %v", error)
	}
	log.Printf("cStream Response %v", response)
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
			log.Printf("sStream Received response: %v", response)
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
			log.Printf("BiDi Received %v", response.GetResult())
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func testAll(client pb.ScorerClient, ctx context.Context) {
	go testUnary(client, ctx)
	go testClientStreaming(client, ctx)
	go testServerStreaming(client, ctx)
	testBiDirectionStreaming(client, ctx)
}
