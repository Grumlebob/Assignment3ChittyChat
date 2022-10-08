package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"time"

	"github.com/Grumlebob/Assignment3ChittyChat/protos"

	"google.golang.org/grpc"
)

type Server struct {
	protos.ChatServiceServer
	messageChannels map[int32]chan *protos.ChatMessage
}

func (s *Server) GetClientId(ctx context.Context, clientMessage *protos.ClientRequest) (*protos.ServerResponse, error) {
	fmt.Println("Server pinged:", time.Now(), "by client:", clientMessage.ChatMessage.Userid)
	//If user exists:
	if s.messageChannels[clientMessage.ChatMessage.Userid] != nil {
		fmt.Println("User exists with ID: ", clientMessage.ChatMessage.Userid)
		return &protos.ServerResponse{
			ChatMessage: &protos.ChatMessage{
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
			s.messageChannels[int32(idgenerator)] = make(chan *protos.ChatMessage)
			break
		}
		idgenerator = rand.Intn(math.MaxInt32)
	}
	fmt.Println("generated new user with ID:", idgenerator)

	return &protos.ServerResponse{
		ChatMessage: &protos.ChatMessage{
			Message:     "Client ID: " + string(idgenerator),
			Userid:      int32(idgenerator),
			LamportTime: 0,
		},
	}, nil
}

func (s *Server) SendMessage(ctx context.Context, clientMessage *protos.ClientRequest, messageStream protos.ChatService_PublishMessageServer) error {
	//broadcast to all channels
	for _, channel := range s.messageChannels {
		channel <- clientMessage.ChatMessage
	}

	messageStream.Send(&protos.ServerResponse{
		ChatMessage: &protos.ChatMessage{
			Message:     "Message sent",
			Userid:      clientMessage.ChatMessage.Userid,
			LamportTime: clientMessage.ChatMessage.LamportTime,
		},
	})

	return nil
}

func main() {
	// Create listener tcp on port 9080

	listener, err := net.Listen("tcp", ":9080")

	if err != nil {
		log.Fatalf("Failed to listen on port 9080: %v", err)
	}

	grpcServer := grpc.NewServer()
	protos.RegisterChatServiceServer(grpcServer, &Server{
		messageChannels: make(map[int32]chan *protos.ChatMessage),
	})

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to server %v", err)
	}

}
