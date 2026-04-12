package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"payment-service/internal/app"
	grpcTransport "payment-service/internal/transport/grpc"

	"github.com/Moldirkab/ap2-generated/paymentpb"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	paymentGRPCPort := os.Getenv("PAYMENT_GRPC_PORT")

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

	application := app.NewApp(db)

	lis, err := net.Listen("tcp", ":"+paymentGRPCPort)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpcTransport.LoggingInterceptor),
	)

	paymentpb.RegisterPaymentServiceServer(server, application.GRPCServer)

	log.Println("payment gRPC listening on", ":"+paymentGRPCPort)
	if err := server.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
