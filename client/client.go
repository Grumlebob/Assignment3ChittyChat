package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Grumlebob/Assignment3ChittyChat/protos"
)

func main() {
	// Create a virtual RPC Client Connection on port 9080
	var conn *grpc.ClientConn
	context, cancelFunction := context.WithTimeout(context.Background(), time.Second*2) //standard er 5
	defer cancelFunction()
	// IPv4:port = "172.30.48.1:9080"
	conn, err := grpc.DialContext(context, "172.30.48.1:9080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	}
	defer conn.Close()

	//  Create new Client from generated gRPC code from proto
	client := protos.NewChatServiceClient(conn)

	sendMessage(client, context)
}

func sendMessage(client protos.ChatServiceClient, context context.Context) {

	fmt.Println("T1:")

	clientRequest := protos.ClientRequest{
		Timestamp: time1,
	}

	response, err := client.GetTime(context, &clientRequest)
	if err != nil {
		log.Fatalf("Error when calling server: %s", err)
	}
	fmt.Println(response)

}
