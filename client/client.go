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
var lamportTime = int64(0)

func main() {
	//Print date and time in log.
	log.SetFlags(log.LstdFlags)
	// Create a virtual RPC Client Connection on port 9080
	var conn *grpc.ClientConn
	context, cancelFunction := context.WithTimeout(context.Background(), time.Second*9999) //standard er 5
	defer cancelFunction()
	//format is IPv4:port eg "172.30.48.1:9080"
	conn, err := grpc.DialContext(context, ":9080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	}
	defer conn.Close()

	//  Create new Client from generated gRPC code from proto
	client := pb.NewChatServiceClient(conn)

	//Blocking, to get client ID
	getClientId(client, context)

	//Leave chat on exit.
	defer leaveChat(client, context)

	/* leave chat with ctrl+c - currently not working.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		leaveChat(client, context)
		fmt.Println("kom her til")
	}()
	*/

	//Non-blocking, to enable client to send messages
	go joinChat(client, context)

	fmt.Println("Enter 'leave()' to leave the chatroom. \nEnter your message here:")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if len(scanner.Text()) > 128 {
			fmt.Println("Message too long. Maximum length is 128 characters.")
		} else {
			go sendMessage(client, context, scanner.Text())
		}
	}
}

func getClientId(client pb.ChatServiceClient, context context.Context) {
	clientRequest := pb.ClientRequest{
		ChatMessage: &pb.ChatMessage{
			Message:     "New User",
			Userid:      userId,
			LamportTime: lamportTime,
		},
	}
	user, err := client.GetClientId(context, &clientRequest)
	if err != nil {
		log.Fatalf("Error when calling server: %s", err)
	}

	userId = user.ChatMessage.Userid
	log.Println("Hello! - You are ID: ", userId)
}

func joinChat(client pb.ChatServiceClient, context context.Context) {
	clientRequest := pb.ClientRequest{
		ChatMessage: &pb.ChatMessage{
			Message:     fmt.Sprint("Participant ", userId, " joined Chitty-Chat"),
			Userid:      userId,
			LamportTime: lamportTime,
		},
	}

	stream, err := client.JoinChat(context, &clientRequest)
	if err != nil {
		log.Fatalf("Error when joining chat server: %s", err)
	}
	joinedString := fmt.Sprintf("Participant %d joined Chitty-Chat", userId)
	sendMessage(client, context, joinedString)

	//Keep them in chatroom until they leave.
	KeepStreamAlive := make(chan struct{})
	go func() {
		for {
			message, err := stream.Recv()
			if err == io.EOF {
				close(KeepStreamAlive)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive message from channel joining. \nErr: %v", err)
			}
			if userId != message.Userid {
				if message.LamportTime > lamportTime {
					lamportTime = message.LamportTime + 1
				} else {
					lamportTime++
				}
			}
			log.Println("User:", message.Userid, "- Lamport time:", lamportTime, "- Msg:", message.Message)
		}
	}()
	<-KeepStreamAlive
}

func sendMessage(client pb.ChatServiceClient, context context.Context, message string) {
	if message == "leave()" {
		leaveChat(client, context)
		return
	}
	lamportTime++

	clientRequest := &pb.ClientRequest{
		ChatMessage: &pb.ChatMessage{
			Message:     message,
			Userid:      userId,
			LamportTime: lamportTime,
		},
	}
	//Handles the response in "JoinChat loop", so just discard here.
	_, err := client.PublishMessage(context, clientRequest)
	if err != nil {
		log.Fatalf("Opening stream: %s", err)
	}
}

func leaveChat(client pb.ChatServiceClient, context context.Context) {

	clientRequest := &pb.ClientRequest{
		ChatMessage: &pb.ChatMessage{
			Message:     "leave()",
			Userid:      userId,
			LamportTime: lamportTime,
		},
	}

	leaveString := fmt.Sprintf("Participant %d left Chitty-Chat", userId)
	sendMessage(client, context, leaveString)

	_, err := client.LeaveChat(context, clientRequest)
	if err != nil {
		log.Fatalf("Error when leaving chat: %s", err)
	}

	//log.Println("Client ", userId, " left chat with response:", reponse.ChatMessage.Message)
	os.Exit(0)
}
