package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"

	"github.com/google/uuid"
	"gitlab.com/keshavbhattad/pcbook/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const MaxImageSize = 2 << 20

type LaptopServer struct {
	laptopStore LaptopStore
	imageStore  ImageStore
	ratingStore RatingStore
}

func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) *LaptopServer {
	return &LaptopServer{
		laptopStore: laptopStore,
		imageStore:  imageStore,
		ratingStore: ratingStore,
	}
}

func (server *LaptopServer) SearchLaptop(
	req *pb.SearchLaptopRequest,
	stream pb.LaptopService_SearchLaptopServer,
) error {
	filter := req.GetFilter()
	log.Printf("Received a search-laptop request with filter: %v", filter)

	err := server.laptopStore.Search(
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

	if err := contextError(ctx); err != nil {
		return nil, err
	}

	// Save laptop Id on database normally
	// Here laptop Id is stored in-memory
	err := server.laptopStore.Save(laptop)
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

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "Request is canceled"))

	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "Deadline is exeeded"))
	default:
		return nil
	}
}

func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "Cannot receive image info"))
	}

	laptopID := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()
	log.Printf("Received an upload-image request for laptop %s with image type %s\n", laptopID, imageType)

	laptop, err := server.laptopStore.Find(laptopID)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "Cannot find laptop: %v", err))
	}
	if laptop == nil {
		return logError(status.Errorf(codes.NotFound, "Laptop not found: %v", err))
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		if err := contextError(stream.Context()); err != nil {
			return nil
		}

		log.Print("Waiting to receive more data")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("No more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "Cannot receive chunk: %v", err))
		}

		chunk := req.GetChunkData()
		size := len(chunk)
		imageSize += size

		if imageSize > MaxImageSize {
			return logError(status.Error(codes.InvalidArgument, "File is too large"))
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "Cannot write chunk data: %v", err))
		}
	}

	imageID, err := server.imageStore.Save(laptopID, imageType, imageData)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "Cannot save image to the store: %v", err))
	}

	res := &pb.UploadImageResponse{
		Id:   imageID,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "Cannot send the response and close the stream: %v", err))
	}

	log.Printf("Image is successfully saved with id: %s and size: %d", imageID, imageSize)

	return nil
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}

func (server *LaptopServer) RateLaptop(stream pb.LaptopService_RateLaptopServer) error {
	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}
		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("No more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "Cannot receive stream request: %v", err))
		}

		laptopID := req.GetLaptopId()
		score := req.GetScore()

		log.Printf("Received a request with laptop ID: %s and score: %.2f", laptopID, score)

		found, err := server.laptopStore.Find(laptopID)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "Cannot find a laptop: %v", err))
		}
		if found == nil {
			return logError(status.Errorf(codes.NotFound, "Laptop with id %s is not found", laptopID))
		}

		rating, err := server.ratingStore.Add(laptopID, score)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "Cannot add the rating: %v", err))
		}

		res := &pb.RateLaptopResponse{
			LaptopId:     laptopID,
			RatedCount:   rating.Count,
			AverageScore: rating.Sum / float64(rating.Count),
		}

		err = stream.Send(res)
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "Cannot send the response and close the stream: %v", err))
		}

		log.Printf("Rating added successfully with rated count: %d and average score: %f", res.GetRatedCount(), res.GetAverageScore())
	}
	return nil
}
