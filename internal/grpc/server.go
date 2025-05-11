package grpc

import (
	"context"
	"log"
	"net"

	pb "github.com/shzuzu/Go_Calculator/pkg/api"
	"github.com/shzuzu/Go_Calculator/pkg/calc"
	"google.golang.org/grpc"
)

type CalculatorServer struct {
	pb.UnimplementedCalculatorServiceServer
}

func (s *CalculatorServer) Calculate(ctx context.Context, req *pb.CalculateRequest) (*pb.CalculateResponse, error) {
	log.Printf("Received calculation request: %s", req.Expression)

	result, err := calc.Calc(req.Expression)
	response := &pb.CalculateResponse{
		Result: result,
	}

	if err != nil {
		response.Error = err.Error()
	}

	return response, nil
}

func (s *CalculatorServer) ValidateExpression(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	log.Printf("Received validation request: %s", req.Expression)

	wp := calc.NewWorkerPool(1) // Create a temporary worker pool for validation
	err := wp.ValidateExpression(req.Expression)

	response := &pb.ValidateResponse{
		IsValid: err == nil,
	}

	if err != nil {
		response.Error = err.Error()
	}

	return response, nil
}

func StartServer(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
		return err
	}

	server := grpc.NewServer()
	pb.RegisterCalculatorServiceServer(server, &CalculatorServer{})

	log.Printf("gRPC server listening on %s", address)
	return server.Serve(lis)
}
