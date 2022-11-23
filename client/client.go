package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	gRPC "github.com/fredpetersen/disys-handin-5/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var servers []gRPC.AuctionHouseClient //the servers
var id int32

func main() {
	for port := 5000; port <= 5002; port++ {
		servers = append(servers, ConnectToServer(port))
	}

	//Get id from each server (hopefully the same :))
	for i := 0; i < 3; i++ {
		Connect(servers[i])
	}
	f, err := os.OpenFile(fmt.Sprintf("client_%v.log",id), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening logfile: %v", err)
	}

	log.SetOutput(f)

	readInput()
}

func readInput() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		in := scanner.Text()
		if in == "exit" { break }
		
		if amount, err := strconv.ParseInt(in,10,32); err == nil {
			Bid(int32(amount), servers)
		} else if in == "result" {
			Result(servers)
		}
		
	}
}

func Connect(server gRPC.AuctionHouseClient) {
	ack, err := server.Bid(context.Background(), &gRPC.BidType{Amount: 0, Id: -1})
	if err != nil {
		log.Fatalf("encountered an error: %v\n", err)
	}
	id = ack.Id;
}

func Bid(amount int32, servers []gRPC.AuctionHouseClient) {
	hasPrinted := false
	for i := 0; i<3; i++ {
		server := servers[i]
		ack, _ := server.Bid(context.Background(), &gRPC.BidType {
			Amount: amount,
			Id:     id,
		});

		if hasPrinted {
			continue
		}

		if ack == nil {
			log.Printf("Received no response from server %d, continuing to try other servers\n", i)
			continue
		}
		
		if (ack.Status == gRPC.BidStatus_SUCCESS) {
			log.Printf("%d successfully bid %d\n", id, amount)
		} else if (ack.Status == gRPC.BidStatus_TOO_LOW) {
			log.Printf("Your bid of %d is lower than the current highest bid\n", amount)
		} else {
			log.Println("The auction has ended! no more bids allowed!!!!",)
		}
		hasPrinted = true;
	}
}

func Result(servers []gRPC.AuctionHouseClient) {
	hasPrinted := false
	for i := 0; i < 3; i++ {
		res, err := servers[i].Result(context.Background(), &emptypb.Empty{});
		if hasPrinted {
			continue
		}
		if err == nil {
			hasPrinted = true
			if (res.Over) {
				log.Printf("The auction is over! %d won with a bid of %d\n", res.Winner, res.Amount)
			} else {
				log.Printf("The current highest bid is %d from client id %d\n", res.Amount, res.Winner)
			}
		}
	}

	if hasPrinted == false {
		log.Printf("All servers have crashed.\n")
	}
}

func ConnectToServer(port int) gRPC.AuctionHouseClient {
	var opts []grpc.DialOption
	var target = fmt.Sprintf("localhost:%d",port)
	opts = append(opts, grpc.WithBlock(), grpc.WithTimeout(1*time.Second), grpc.WithTransportCredentials(insecure.NewCredentials()))
	fmt.Printf("Dialing on %s \n", target)
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		log.Fatalf("Fail to dial: %v\n", err)
	}

	return gRPC.NewAuctionHouseClient(conn)
}