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

	//Blocking, to get client ID
	getClientId(client, context)

	//Enables client to leave with ctrl+c
	defer leaveChat(client, context)

	//Non-blocking, to enable client to send messages
	go joinChat(client, context)

	fmt.Println("Enter 'leave()' to leave the chatroom or use hotkey ctrl+c \nEnter your message here:")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		go sendMessage(client, context, scanner.Text())
		fmt.Println("Please enter message:")
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
			log.Println("User: ", message.Userid, " - ", message.Message, "at Lamport time: ", message.LamportTime)
		}
	}()
	<-loopForever
}

func sendMessage(client pb.ChatServiceClient, context context.Context, message string) {
	if message == "leave()" {
		leaveChat(client, context)
		return
	}

	clientRequest := &pb.ClientRequest{
		ChatMessage: &pb.ChatMessage{
			Message:     message,
			Userid:      userId,
			LamportTime: 0,
		},
	}
	//Handles the response in "JoinChat loop"
	_, err := client.PublishMessage(context, clientRequest)
	if err != nil {
		log.Fatalf("Opening stream: %s", err)
	}
}

func leaveChat(client pb.ChatServiceClient, context context.Context) {
	fmt.Println("Client ", userId, " attempts to leave chat")

	clientRequest := &pb.ClientRequest{
		ChatMessage: &pb.ChatMessage{
			Message:     "leave()",
			Userid:      userId,
			LamportTime: 0,
		},
	}

	response, err := client.LeaveChat(context, clientRequest)
	if err != nil {
		log.Fatalf("Error when leaving chat: %s", err)
	}
	fmt.Println("Client ", userId, " left chat: ", response.ChatMessage.Message)
	os.Exit(0)
}
