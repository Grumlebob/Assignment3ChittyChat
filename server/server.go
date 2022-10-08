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
	fmt.Println("Server pinged:", time.Now())
	idgenerator := rand.Intn(math.MaxInt32)
	fmt.Println("Random number:", idgenerator)
	for {
		if s.messageChannels[int32(idgenerator)] == nil {
			s.messageChannels[int32(idgenerator)] = make(chan *protos.ChatMessage)
			s.messageChannels[int32(idgenerator)] <- clientMessage.ChatMessage
			break
		}
		idgenerator = rand.Intn(math.MaxInt32)
	}
	fmt.Println("Out of id loop:", idgenerator)

	return &protos.ServerResponse{
		ChatMessage: &protos.ChatMessage{
			Message:     "Client ID: " + string(idgenerator),
			Userid:      int32(idgenerator),
			LamportTime: 0,
		},
	}, nil
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
