package serializer

import (
	"fmt"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
)

func WriteProtobufToJSONFile(message proto.Message, fileName string) error {
	data, err := ProtobufToJSON(message)
	if err != nil {
		return fmt.Errorf("Cannot marshal proto message to JSON: %w", err)
	}

	err = ioutil.WriteFile(fileName, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("Cannot write to JSON file: %w", err)
	}

	return nil
}

func WriteProtobufToBinaryFile(message proto.Message, fileName string) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("Cannot marshal proto message to binary: %w", err)
	}

	err = ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		return fmt.Errorf("Cannot write to binary file: %w", err)
	}

	return nil
}

func ReadProtobufFromBinaryFile(fileName string, message proto.Message) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("Cannot read from binary file: %w", err)
	}

	err = proto.Unmarshal(data, message)
	if err != nil {
		return fmt.Errorf("Cannot unmarshal binary to proto message: %w", err)
	}

	return nil
}
