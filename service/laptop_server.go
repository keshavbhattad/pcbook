package service

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"gitlab.com/keshavbhattad/pcbook/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LaptopServer struct {
	Store LaptopStore
}

func NewLaptopServer(store LaptopStore) *LaptopServer {
	return &LaptopServer{
		Store: store,
	}
}

func (server *LaptopServer) SearchLaptop(
	req *pb.SearchLaptopRequest,
	stream pb.LaptopService_SearchLaptopServer,
) error {
	filter := req.GetFilter()
	log.Printf("Received a search-laptop request with filter: %v", filter)

	err := server.Store.Search(
		stream.Context(),
		filter,
		func(laptop *pb.Laptop) error {
			res := &pb.SearchLaptopResponse{Laptop: laptop}

			err := stream.Send(res)

			if err != nil {
				return err
			}

			log.Printf("sent laptop with ID: %s", laptop.GetId())
			return nil
		},
	)

	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	return nil
}

func (server *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("Received create-laptop request with ID: %s", laptop.Id)

	if len(laptop.Id) > 0 {
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Laptop ID is not valid UUID: %v", err)
		}

	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Cannot generate a new Laptop ID: %v", err)
		}
		laptop.Id = id.String()
	}

	if ctx.Err() == context.Canceled {
		log.Print("Request is canceled")
		return nil, status.Error(codes.Canceled, "Request is canceled")
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Print("Deadline is exceeded")
		return nil, status.Error(codes.DeadlineExceeded, "Deadline is exeeded")
	}

	// Save laptop Id on database normally
	// Here laptop Id is stored in-memory
	err := server.Store.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExists) {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "Cannot save laptop to the store: %v", err)
	}
	log.Printf("Laptop saved with id: %s", laptop.Id)
	res := &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}
	return res, nil
}
