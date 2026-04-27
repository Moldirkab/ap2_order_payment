package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Moldirkab/ap2-generated/ordertrackingpb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter order ID: ")
	orderID, _ := reader.ReadString('\n')
	orderID = strings.TrimSpace(orderID)

	if orderID == "" {
		log.Fatal("order ID cannot be empty")
	}
	addr := os.Getenv("ORDER_TRACKING_ADDR")
	if addr == "" {
		log.Fatal("ORDER_TRACKING_ADDR is required")
	}
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("failed to connect:", err)
	}
	defer conn.Close()

	client := ordertrackingpb.NewOrderTrackingServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	stream, err := client.SubscribeToOrderUpdates(ctx, &ordertrackingpb.OrderRequest{
		OrderId: orderID,
	})
	if err != nil {
		log.Fatal("stream error:", err)
	}

	fmt.Println("\nListening for updates...\n")

	for {
		update, err := stream.Recv()
		if err != nil {
			log.Println("stream closed:", err)
			return
		}

		fmt.Println("----- ORDER UPDATE -----")
		fmt.Println("ID:     ", update.OrderId)
		fmt.Println("Status: ", update.Status)
		fmt.Println("Amount: ", update.Amount)

		if update.CreatedAt != nil {
			fmt.Println("Created:", update.CreatedAt.AsTime().Format(time.RFC3339))
		}

		fmt.Println("------------------------\n")
	}
}
