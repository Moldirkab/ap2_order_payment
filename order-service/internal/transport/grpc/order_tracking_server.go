package grpc

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"order-service/internal/repository/postgres"

	"github.com/Moldirkab/ap2-generated/ordertrackingpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderTrackingServer struct {
	ordertrackingpb.UnimplementedOrderTrackingServiceServer
	db *sql.DB
}

func NewOrderTrackingServer(db *sql.DB) *OrderTrackingServer {
	return &OrderTrackingServer{db: db}
}

func (s *OrderTrackingServer) SubscribeToOrderUpdates(
	req *ordertrackingpb.OrderRequest,
	stream ordertrackingpb.OrderTrackingService_SubscribeToOrderUpdatesServer,
) error {
	repo := postgres.NewOrderRepository(s.db)

	orderID := req.GetOrderId()
	if orderID == "" {
		return errors.New("order_id is required")
	}

	lastStatus := ""

	sendCurrent := func(ctx context.Context) error {
		order, err := repo.GetByID(orderID)
		if err != nil {
			return err
		}

		if order.Status != lastStatus {
			lastStatus = order.Status
			return stream.Send(&ordertrackingpb.OrderStatusUpdate{
				OrderId:   order.ID,
				Status:    order.Status,
				Amount:    order.Amount,
				CreatedAt: timestamppb.New(order.CreatedAt),
			})
		}

		return nil
	}

	if err := sendCurrent(stream.Context()); err != nil {
		return err
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			if err := sendCurrent(stream.Context()); err != nil {
				return err
			}
		}
	}
}
