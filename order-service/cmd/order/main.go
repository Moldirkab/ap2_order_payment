package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"order-service/internal/app"

	"github.com/Moldirkab/ap2-generated/ordertrackingpb"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	_ = godotenv.Load()

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	orderHTTPPort := os.Getenv("ORDER_HTTP_PORT")
	orderGRPCPort := os.Getenv("ORDER_GRPC_PORT")
	paymentGRPCAddr := os.Getenv("PAYMENT_GRPC_ADDR")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("db not reachable")
	}

	application, err := app.NewApp(db, paymentGRPCAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer application.PaymentClient.Close()

	go func() {
		r := gin.Default()
		application.HTTPHandler.RegisterRoutes(r)

		addr := ":" + orderHTTPPort
		log.Println("order REST listening on", addr)
		if err := r.Run(addr); err != nil {
			log.Fatal(err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+orderGRPCPort)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	ordertrackingpb.RegisterOrderTrackingServiceServer(grpcServer, application.OrderTrackingServer)

	log.Println("order gRPC listening on", ":"+orderGRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
