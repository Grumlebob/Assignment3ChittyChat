package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"time"

	pb "github.com/Grumlebob/Assignment3ChittyChat/protos"

	"google.golang.org/grpc"
)

type Server struct {
	pb.ChatServiceServer
	messageChannels map[int32]chan *pb.ChatMessage
}

func main() {
	// Create listener tcp on port 9080
	listener, err := net.Listen("tcp", ":9080")
	if err != nil {
		log.Fatalf("Failed to listen on port 9080: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterChatServiceServer(grpcServer, &Server{
		messageChannels: make(map[int32]chan *pb.ChatMessage),
	})
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to server %v", err)
	}
}

func (s *Server) GetClientId(ctx context.Context, clientMessage *pb.ClientRequest) (*pb.ServerResponse, error) {
	fmt.Println("Server pinged:", time.Now(), "by client:", clientMessage.ChatMessage.Userid)
	//If user exists:
	if s.messageChannels[clientMessage.ChatMessage.Userid] != nil {
		fmt.Println("User exists with ID: ", clientMessage.ChatMessage.Userid)
		return &pb.ServerResponse{
			ChatMessage: &pb.ChatMessage{
				Message:     clientMessage.ChatMessage.Message,
				Userid:      clientMessage.ChatMessage.Userid,
				LamportTime: clientMessage.ChatMessage.LamportTime,
			},
		}, nil
	}
	//If user doesn't exist:
	idgenerator := rand.Intn(math.MaxInt32)
	for {
		if s.messageChannels[int32(idgenerator)] == nil {
			break
		}
		idgenerator = rand.Intn(math.MaxInt32)
	}
	fmt.Println("generated new user with ID:", idgenerator)

	return &pb.ServerResponse{
		ChatMessage: &pb.ChatMessage{
			Message:     "Client ID: " + string(idgenerator),
			Userid:      int32(idgenerator),
			LamportTime: 0,
		},
	}, nil
}

// rpc ListFeatures(Rectangle) returns (stream Feature) {} eksempelt. A server-side streaming RPC
func (s *Server) PublishMessage(clientMessage *pb.ClientRequest, stream pb.ChatService_PublishMessageServer) error {
	fmt.Println("Server publishish message from user: ", clientMessage.ChatMessage.Userid, "Message:", clientMessage.ChatMessage.Message)

	if s.messageChannels[clientMessage.ChatMessage.Userid] == nil {
		s.messageChannels[clientMessage.ChatMessage.Userid] = make(chan *pb.ChatMessage)
		fmt.Println("Added user stream to map.", clientMessage.ChatMessage.Userid)
	}

	response := &pb.ServerResponse{
		ChatMessage: &pb.ChatMessage{
			Message:     "Message sent: " + clientMessage.ChatMessage.Message,
			Userid:      clientMessage.ChatMessage.Userid,
			LamportTime: clientMessage.ChatMessage.LamportTime,
		},
	}

	//broadcast to all channels
	fmt.Println("enter broadcasting")
	totalUsers := len(s.messageChannels)
	fmt.Println("Total users: ", len(s.messageChannels))
	for _, channels := range s.messageChannels {
		totalUsers--
		channels <- response.ChatMessage
		if totalUsers == 0 {
			break
		}
	}
	fmt.Println("Left broadcasting")
	return nil
}

// rpc ListFeatures(Rectangle) returns (stream Feature) {} eksempelt. A server-side streaming RPC
func (s *Server) JoinChat(clientMessage *pb.ClientRequest, stream pb.ChatService_JoinChatServer) error {
	fmt.Println("User joined chat: ", clientMessage.ChatMessage.Userid)

	if s.messageChannels[clientMessage.ChatMessage.Userid] == nil {
		messageChannel := make(chan *pb.ChatMessage)
		s.messageChannels[clientMessage.ChatMessage.Userid] = messageChannel
		fmt.Println("Added user chan to map.", clientMessage.ChatMessage.Userid)
	}

	////Keep them in chatroom until they leave.
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case message := <-s.messageChannels[clientMessage.ChatMessage.Userid]:
			stream.Send(message)
		}
	}
}
