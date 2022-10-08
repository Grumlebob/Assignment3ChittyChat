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

var clientId int32

func main() {
	// Create a virtual RPC Client Connection on port 9080
	var conn *grpc.ClientConn
	context, cancelFunction := context.WithTimeout(context.Background(), time.Second*200) //standard er 5
	defer cancelFunction()
	// IPv4:port = "172.30.48.1:9080"
	conn, err := grpc.DialContext(context, ":9080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	}
	defer conn.Close()

	//  Create new Client from generated gRPC code from proto
	client := protos.NewChatServiceClient(conn)

	getClientId(client, context)
	//sendMessage(client, context)
}

func getClientId(client protos.ChatServiceClient, context context.Context) {
	clientRequest := protos.ClientRequest{
		ChatMessage: &protos.ChatMessage{
			Message:     "New User",
			Userid:      clientId,
			LamportTime: 0,
		},
	}
	user, err := client.GetClientId(context, &clientRequest)
	if err != nil {
		log.Fatalf("Error when calling server: %s", err)
	}

	clientId = user.ChatMessage.Userid
	fmt.Println("Hello! - You are ID: ", clientId)
}

func sendMessage(client protos.ChatServiceClient, context context.Context) {

	fmt.Println("Client sends message: ")

	clientRequest := protos.ClientRequest{
		ChatMessage: &protos.ChatMessage{
			Message:     "Hello World",
			Userid:      5,
			LamportTime: 0,
		},
	}

	steamOfResponses, err := client.PublishMessage(context, &clientRequest)
	if err != nil {
		log.Fatalf("Error when calling server: %s", err)
	}

	steamOfResponses.SendMsg(clientRequest)

	for {
		response, err := steamOfResponses.Recv()
		if err != nil {
			log.Fatalf("Error when receiving response from server: %s", err)
		}
		fmt.Println("Response from server: ", response)
	}

}
