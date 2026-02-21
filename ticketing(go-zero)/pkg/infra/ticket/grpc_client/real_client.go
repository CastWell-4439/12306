package grpc_client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	seatallocatorv1 "ticketing-gozero/spec/proto/seatallocator/v1"
)

type GRPCSeatAllocator struct {
	conn       *grpc.ClientConn
	client     seatallocatorv1.SeatAllocatorClient
	trainID    string
	travelDate string
	coachType  string
	fromIndex  int
	toIndex    int
}

func NewGRPCSeatAllocator(
	addr string,
	trainID string,
	travelDate string,
	coachType string,
	fromIndex int,
	toIndex int,
) (*GRPCSeatAllocator, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &GRPCSeatAllocator{
		conn:       conn,
		client:     seatallocatorv1.NewSeatAllocatorClient(conn),
		trainID:    trainID,
		travelDate: travelDate,
		coachType:  coachType,
		fromIndex:  fromIndex,
		toIndex:    toIndex,
	}, nil
}

func (c *GRPCSeatAllocator) AllocateSeat(ctx context.Context, orderID string) (string, error) {
	resp, err := c.client.AllocateSeat(ctx, &seatallocatorv1.AllocateSeatRequest{
		OrderId:    orderID,
		TrainId:    c.trainID,
		TravelDate: c.travelDate,
		CoachType:  c.coachType,
		FromIndex:  uint32(c.fromIndex),
		ToIndex:    uint32(c.toIndex),
	})
	if err != nil {
		return "", err
	}
	return resp.GetSeatNo(), nil
}

func (c *GRPCSeatAllocator) Close() error {
	return c.conn.Close()
}


