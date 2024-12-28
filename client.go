package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"multiplayer-webservice/internal/proto" // Update to the correct import path
)

func main() {
	// Set up a connection to the server
	conn, err := grpc.Dial(":50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Create a new client for MultiplayerService
	client := proto.NewMultiplayerServiceClient(conn)

	// Call the GetModeUsage method
	req := &proto.ModeUsageRequest{
		AreaCode: "12345", // Use a test AreaCode
	}

	res, err := client.GetModeUsage(context.Background(), req)
	if err != nil {
		log.Fatalf("could not get mode usage: %v", err)
	}

	// Print the response from the server
	fmt.Println("Response:")
	for _, mode := range res.Modes {
		fmt.Printf("Mode: %s, Active Users: %d\n", mode.ModeName, mode.ActiveUsers)
	}
}
