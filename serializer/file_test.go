package serializer_test

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	"gitlab.com/keshavbhattad/pcbook/pb"
	"gitlab.com/keshavbhattad/pcbook/sample"
	"gitlab.com/keshavbhattad/pcbook/serializer"
)

func TestFileSerializer(t *testing.T) {
	t.Parallel()

	binaryFile := "../tmp/laptop.bin"

	laptop1 := sample.NewLaptop()
	err := serializer.WriteProtobufToBinaryFile(laptop1, binaryFile)
	require.NoError(t, err)

	laptop2 := &pb.Laptop{}
	err = serializer.ReadProtobufFromBinaryFile(binaryFile, laptop2)
	require.NoError(t, err)
	require.True(t, proto.Equal(laptop1, laptop2))

	JSONFile := "../tmp/laptop.json"

	err = serializer.WriteProtobufToJSONFile(laptop1, JSONFile)
	require.NoError(t, err)
}
