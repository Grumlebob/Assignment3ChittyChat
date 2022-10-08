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
	messageChannels map[int32]chan *pb.ChatMessage //Lav om til slice.
	streams         []*pb.ChatService_PublishMessageServer
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
			s.messageChannels[int32(idgenerator)] = make(chan *pb.ChatMessage, 1)
			break
		}
		idgenerator = rand.Intn(math.MaxInt32)
	}
	fmt.Println("generated new user with ID:", idgenerator)
	fmt.Println("Total users: ", len(s.messageChannels))

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
	if s.streams[clientMessage.ChatMessage.Userid] == nil {
		fmt.Println("gik ind i første loop, godt.")
		s.streams = append(s.streams, &stream)
	}
	fmt.Println("Server trying to publish message from user: ", clientMessage.ChatMessage.Userid)
	response := &pb.ServerResponse{
		ChatMessage: &pb.ChatMessage{
			Message:     "Message sent: " + clientMessage.ChatMessage.Message,
			Userid:      clientMessage.ChatMessage.Userid,
			LamportTime: clientMessage.ChatMessage.LamportTime,
		},
	}
	if err := stream.Send(response); err != nil {
		log.Printf("send error %v", err)
	}
	//broadcast to all channels
	fmt.Println("enter broadcasting:")
	/*
		for i, channel := range s.messageChannels {
			fmt.Println("broadcasting to channel: ", i)
			channel <- response.ChatMessage
			fmt.Println("i loop broadcasting:")
		}
	*/
	return nil
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
		streams:         make([]*pb.ChatService_PublishMessageServer, 10),
	})
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to server %v", err)
	}

}
