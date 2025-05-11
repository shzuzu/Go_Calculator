package grpc

import (
	"context"
	"errors"
	"log"
	"time"

	pb "github.com/shzuzu/Go_Calculator/pkg/api"
	"github.com/shzuzu/Go_Calculator/pkg/calc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CalculatorClient struct {
	client pb.CalculatorServiceClient
	conn   *grpc.ClientConn
}

func NewCalculatorClient(address string) (*CalculatorClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
		return nil, err
	}

	client := pb.NewCalculatorServiceClient(conn)
	return &CalculatorClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *CalculatorClient) Close() error {
	return c.conn.Close()
}

func (c *CalculatorClient) Calculate(expression string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := c.client.Calculate(ctx, &pb.CalculateRequest{
		Expression: expression,
	})

	if err != nil {
		log.Printf("Failed to calculate expression: %v", err)
		return 0, err
	}

	if response.Error != "" {
		return 0, errors.New(response.Error)
	}

	return response.Result, nil
}

func (c *CalculatorClient) ValidateExpression(expression string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := c.client.ValidateExpression(ctx, &pb.ValidateRequest{
		Expression: expression,
	})

	if err != nil {
		log.Printf("Failed to validate expression: %v", err)
		return err
	}

	if !response.IsValid {
		if response.Error == "division by zero" {
			return calc.ErrDivisionByZero
		} else if response.Error == "EOF" {
			return calc.ErrEOF
		} else {
			return calc.ErrInvalidExpression
		}
	}

	return nil
}
