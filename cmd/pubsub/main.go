package main

import (
	"flag"

	"google.golang.org/grpc"
	"log"
	"net"

	"github.com/real-nil/grpc-pubsub/proto/pubsub"
)

var (
	port = flag.String(`addr`, `:3456`, `address of pubsub server`)
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pubsub.RegisterPubSubServer(s, nil)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
