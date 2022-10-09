package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"

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
	//fmt.Println("generated new user with ID:", idgenerator)

	return &pb.ServerResponse{
		ChatMessage: &pb.ChatMessage{
			Message:     "Client ID: " + string(idgenerator),
			Userid:      int32(idgenerator),
			LamportTime: clientMessage.ChatMessage.LamportTime,
		},
	}, nil
}

// rpc ListFeatures(Rectangle) returns (stream Feature) {} eksempelt. A server-side streaming RPC
func (s *Server) PublishMessage(ctx context.Context, clientMessage *pb.ClientRequest) (*pb.ServerResponse, error) {

	if s.messageChannels[clientMessage.ChatMessage.Userid] == nil {
		s.messageChannels[clientMessage.ChatMessage.Userid] = make(chan *pb.ChatMessage)
	}

	response := &pb.ServerResponse{
		ChatMessage: &pb.ChatMessage{
			Message:     clientMessage.ChatMessage.Message,
			Userid:      clientMessage.ChatMessage.Userid,
			LamportTime: clientMessage.ChatMessage.LamportTime,
		},
	}

	//broadcast to all channels
	for _, channels := range s.messageChannels {
		channels <- response.ChatMessage
	}
	return response, nil
}

// rpc ListFeatures(Rectangle) returns (stream Feature) {} eksempelt. A server-side streaming RPC
func (s *Server) JoinChat(clientMessage *pb.ClientRequest, stream pb.ChatService_JoinChatServer) error {
	//If user doesn't have a channel
	if s.messageChannels[clientMessage.ChatMessage.Userid] == nil {
		s.messageChannels[clientMessage.ChatMessage.Userid] = make(chan *pb.ChatMessage)
	}

	log.Println(clientMessage.ChatMessage.Message, " Total users: ", len(s.messageChannels))

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

func (s *Server) LeaveChat(ctx context.Context, clientMessage *pb.ClientRequest) (*pb.ServerResponse, error) {
	//Remove map entry for user
	delete(s.messageChannels, clientMessage.ChatMessage.Userid)
	log.Println("Participant", clientMessage.ChatMessage.Userid, "left Chitty-Chat.", "Total users: ", len(s.messageChannels))

	return &pb.ServerResponse{
		ChatMessage: &pb.ChatMessage{
			Message:     clientMessage.ChatMessage.Message,
			Userid:      clientMessage.ChatMessage.Userid,
			LamportTime: clientMessage.ChatMessage.LamportTime,
		},
	}, nil
}
