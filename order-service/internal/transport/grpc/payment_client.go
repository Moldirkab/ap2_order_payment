package grpc

import (
	"context"
	"time"

	"github.com/Moldirkab/ap2-generated/paymentpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentGRPCClient struct {
	client paymentpb.PaymentServiceClient
	conn   *grpc.ClientConn
}

func NewPaymentGRPCClient(addr string) (*PaymentGRPCClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := paymentpb.NewPaymentServiceClient(conn)

	return &PaymentGRPCClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *PaymentGRPCClient) ProcessPayment(orderID string, amount int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := c.client.ProcessPayment(ctx, &paymentpb.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	})
	if err != nil {
		return "", err
	}

	return resp.Status, nil
}

func (c *PaymentGRPCClient) Close() error {
	return c.conn.Close()
}
