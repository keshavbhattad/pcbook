package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/keshavbhattad/pcbook/pb"
	"gitlab.com/keshavbhattad/pcbook/sample"
	"gitlab.com/keshavbhattad/pcbook/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopNoID := sample.NewLaptop()
	laptopNoID.Id = ""

	laptopInvalidID := sample.NewLaptop()
	laptopInvalidID.Id = "invalid-uuid"

	laptopDuplicateID := sample.NewLaptop()
	storeDuplicateID := service.NewInMemoryLaptopStore()
	err := storeDuplicateID.Save(laptopDuplicateID)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		laptop      *pb.Laptop
		laptopStore service.LaptopStore
		imageStore  service.ImageStore
		code        codes.Code
	}{
		{
			name:        "Success_with_id",
			laptop:      sample.NewLaptop(),
			laptopStore: service.NewInMemoryLaptopStore(),
			imageStore:  nil,
			code:        codes.OK,
		},
		{
			name:        "Success_no_id",
			laptop:      laptopNoID,
			laptopStore: service.NewInMemoryLaptopStore(),
			imageStore:  nil,
			code:        codes.OK,
		},
		{
			name:        "Failure_invalid_uuid",
			laptop:      laptopInvalidID,
			laptopStore: service.NewInMemoryLaptopStore(),
			imageStore:  nil,
			code:        codes.InvalidArgument,
		},
		{
			name:        "Failure_duplicate_id",
			laptop:      laptopDuplicateID,
			laptopStore: storeDuplicateID,
			imageStore:  nil,
			code:        codes.AlreadyExists,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := &pb.CreateLaptopRequest{
				Laptop: tc.laptop,
			}

			server := service.NewLaptopServer(tc.laptopStore, tc.imageStore)
			res, err := server.CreateLaptop(context.Background(), req)

			if tc.code == codes.OK {
				require.NoError(t, err)
				require.NotEmpty(t, res)
				require.NotEmpty(t, res.Id)
				if len(tc.laptop.Id) > 0 {
					require.Equal(t, res.Id, tc.laptop.Id)
				}
			} else {
				require.Error(t, err)
				require.Nil(t, res)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, st.Code(), tc.code)
			}
		})

	}
}
