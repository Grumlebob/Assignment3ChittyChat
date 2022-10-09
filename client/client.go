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

	pb "github.com/Grumlebob/Assignment3ChittyChat/protos"
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
	client := pb.NewChatServiceClient(conn)

	getClientId(client, context)
	//Indsæt join chat methods

	go joinChat(client, context)

	sendMessage(client, context, "hello 1")
	//sendMessage(client, context, "hello 2")
	//sendMessage(client, context, "hello 3")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		sendMessage(client, context, scanner.Text())
	}
}

func getClientId(client pb.ChatServiceClient, context context.Context) {
	clientRequest := pb.ClientRequest{
		ChatMessage: &pb.ChatMessage{
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

func joinChat(client pb.ChatServiceClient, context context.Context) {
	clientRequest := pb.ClientRequest{
		ChatMessage: &pb.ChatMessage{
			Message:     "New User",
			Userid:      userId,
			LamportTime: 0,
		},
	}

	stream, err := client.JoinChat(context, &clientRequest)
	if err != nil {
		log.Fatalf("Error when joining chat server: %s", err)
	}
	fmt.Println("User ", clientRequest.ChatMessage.Userid, " joined the chat")

	//Keep them in chatroom until they leave.
	loopForever := make(chan struct{})
	go func() {
		for {
			message, err := stream.Recv()
			if err == io.EOF {
				close(loopForever)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive message from channel joining. \nErr: %v", err)
			}
			log.Println(message)
		}
	}()
	<-loopForever

}

func sendMessage(client pb.ChatServiceClient, context context.Context, message string) {
	fmt.Println("Client ", userId, " attempts to send message: ", message)

	clientRequest := &pb.ClientRequest{
		ChatMessage: &pb.ChatMessage{
			Message:     message,
			Userid:      userId,
			LamportTime: 0,
		},
	}

	stream, err := client.PublishMessage(context, clientRequest)
	if err != nil {
		log.Fatalf("Opening stream: %s", err)
	}
	for {
		message, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.PublishMessage(_) = _, %v", client, err)
		}
		log.Println(message)
	}
}
