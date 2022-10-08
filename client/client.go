package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Grumlebob/Assignment3ChittyChat/protos"
)

var userId int32

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

	sendMessage(client, context, "hello 1")
	sendMessage(client, context, "hello 2")
	sendMessage(client, context, "hello 3")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		go sendMessage(client, context, scanner.Text())
	}
}

func getClientId(client protos.ChatServiceClient, context context.Context) {
	clientRequest := protos.ClientRequest{
		ChatMessage: &protos.ChatMessage{
			Message:     "New User",
			Userid:      userId,
			LamportTime: 0,
		},
	}
	user, err := client.GetClientId(context, &clientRequest)
	if err != nil {
		log.Fatalf("Error when calling server: %s", err)
	}

	userId = user.ChatMessage.Userid
	fmt.Println("Hello! - You are ID: ", userId)
}

func sendMessage(client protos.ChatServiceClient, context context.Context, message string) {
	fmt.Println("Client ", userId, " attempts to send message: ", message)

	clientRequest := protos.ClientRequest{
		ChatMessage: &protos.ChatMessage{
			Message:     message,
			Userid:      userId,
			LamportTime: 0,
		},
	}

	steamOfResponses, err := client.PublishMessage(context, &clientRequest)
	if err != nil {
		log.Fatalf("Opening stream: %s", err)
	}

	done := make(chan bool)
	go func() {
		for {
			resp, err := steamOfResponses.Recv()
			if err == io.EOF {
				done <- true //close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			fmt.Println("Resp received:", resp.ChatMessage.Message)
		}
	}()
	<-done
	log.Printf("finished")

	//steamOfResponses.SendMsg(&clientRequest)

	/*
		for {
			response, err := steamOfResponses.Recv()
			if err != nil {
				log.Fatalf("Error when receiving response from server: %s", err)
			}
			fmt.Println("Response from server: ", response)
		}
	*/
}
