package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	gRPC "github.com/fredpetersen/disys-handin-5/grpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	gRPC.UnimplementedAuctionHouseServer

	auctionDiesAt				time.Time
	nextValidId					int32
	previousHighestBid 	int32
	highestBidderId 		int32
	port     						string
}

var port = flag.String("port", "5400", "Server port")
var auctionDuration = flag.Duration("duration", 1*time.Minute, "The length of the auction")

func main() {
	flag.Parse()
	listener, _ := net.Listen("tcp", "localhost:"+*port)

	f, err := os.OpenFile(fmt.Sprintf("server_%v.log",*port), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening logfile: %v", err)
	}

	log.SetOutput(f)

	log.Printf("Listening on port %s \n", *port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create server
	grpcServer := grpc.NewServer()
	s := &Server{
		port:     *port,
		auctionDiesAt: time.Now().Add(*auctionDuration),
		nextValidId: 1,
	}
	gRPC.RegisterAuctionHouseServer(grpcServer, s)

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		log.Printf("CRASHED, FULLY EXPECTEDLY. GO PANIC() AAAA\n")
		os.Exit(1)
	}()

	grpcServer.Serve(listener)
}

func (s *Server) Bid(ctx context.Context, bid *gRPC.BidType) (*gRPC.Ack, error) {
	currBid := bid.Amount
	log.Printf("Got a bid with amount %d from  client %d\n", bid.Amount, bid.Id)
	if bid.Id == -1 {
		log.Printf("Bidder has no ID - assigning the id: %d", s.nextValidId)
		out := &gRPC.Ack{Status: gRPC.BidStatus_SUCCESS, Id: s.nextValidId}
		s.nextValidId++
		return out, nil 
	}
	if currBid <= s.previousHighestBid {
		log.Printf("Bid from client %d is too low", bid.Id)
		return &gRPC.Ack{Status: gRPC.BidStatus_TOO_LOW, Id: bid.Id}, nil
	}

	if s.auctionDiesAt.Before(time.Now()) {
		log.Printf("Bid was sent after auction ended")
		return &gRPC.Ack{Status: gRPC.BidStatus_OVER, Id: bid.Id}, nil
	}

	s.previousHighestBid = currBid
	s.highestBidderId = bid.Id
	log.Printf("Bid of %d from client %d was accepted\n", bid.Amount, bid.Id)
	return &gRPC.Ack{Status: gRPC.BidStatus_SUCCESS, Id: bid.Id}, nil

}

func (s *Server) Result(ctx context.Context, _ *emptypb.Empty) (*gRPC.Outcome, error) {
	return &gRPC.Outcome{ Amount: s.previousHighestBid, Over: s.auctionDiesAt.Before(time.Now()), Winner: s.highestBidderId }, nil
}