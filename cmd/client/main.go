package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"gitlab.com/keshavbhattad/pcbook/pb"
	"gitlab.com/keshavbhattad/pcbook/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func createLaptop(laptopClient pb.LaptopServiceClient, laptop *pb.Laptop) {

	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := laptopClient.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Print("Laptop already exists")
		} else {
			log.Fatal("Cannot create laptop: ", err)
		}
		return
	}
	log.Printf("Created laptop with ID: %s", res.Id)
}

func searchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	log.Printf("Search filter : %v", filter)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*8)
	defer cancel()

	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("Cannot search laptop: ", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("Cannot receive response: ", err)
		}

		laptop := res.GetLaptop()
		log.Print(" - found: ", laptop.GetId())
		log.Print(" + brand: ", laptop.GetBrand())
		log.Print(" + name: ", laptop.GetName())
		log.Print(" + CPU cores: ", laptop.Cpu.GetNumberOfCores())
		log.Print(" + CPU minimum GHz: ", laptop.Cpu.GetMinGhz())
	}
}

func testSearchLaptop(laptopClient pb.LaptopServiceClient) {
	for i := 0; i < 10; i++ {
		createLaptop(laptopClient, sample.NewLaptop())
	}

	filter := &pb.Filter{
		MaxPriceInr: 100000,
		MinCpuCores: 4,
		MinCpuGhz:   1.5,
		MinRam: &pb.Memory{
			Value: 2,
			Unit:  pb.Memory_GIGYBYTE,
		},
	}

	searchLaptop(laptopClient, filter)
}

func testCreateLaptop(laptopClient pb.LaptopServiceClient) {
	createLaptop(laptopClient, sample.NewLaptop())
}

func uploadImage(laptopClient pb.LaptopServiceClient, laptopID string, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("Cannot read the file: ", err)
		return
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.UploadImage(ctx)
	if err != nil {
		log.Fatal("cannot upload image: ", err)
	}

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptopID,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("Cannot send the image info: ", err)
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("Cannot read the image file: ", err)
		}

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)

		if err != nil {
			err2 := stream.RecvMsg(nil)
			log.Fatal("Cannot send the image file: ", err, err2)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("Cannot receive the response: ", err)
	}

	log.Printf("Image is uploaded with ID: %s and size: %d", res.GetId(), res.GetSize())
}

func testUploadImage(laptopClient pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	createLaptop(laptopClient, laptop)
	uploadImage(laptopClient, laptop.GetId(), "tmp/image.jpeg")
}

func main() {
	serverAddress := flag.String("address", "", "the server address")
	flag.Parse()
	log.Printf("Dial server at address: %s", *serverAddress)

	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Cannot dial the server: ", err)
	}

	laptopClient := pb.NewLaptopServiceClient(conn)
	testUploadImage(laptopClient)
	// testSearchLaptop(laptopClient)
}
