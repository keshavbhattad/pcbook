package service

import (
	"errors"
	"fmt"
	"sync"

	"github.com/jinzhu/copier"
	"gitlab.com/keshavbhattad/pcbook/pb"
)

var ErrAlreadyExists = errors.New("Record already exists")

type LaptopStore interface {
	Save(laptop *pb.Laptop) error
	Find(id string) (*pb.Laptop, error)
}

type InMemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Laptop
}

func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return fmt.Errorf("Cannot copy laptop data: %w", err)
	}

	store.data[other.Id] = other
	return nil
}

func (store *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop := store.data[id]

	if store.data[id] != nil {
		other := &pb.Laptop{}
		err := copier.Copy(other, laptop)
		if err != nil {
			return nil, fmt.Errorf("Cannot be copied %w", err)
		}
		return other, nil
	} else {
		return nil, fmt.Errorf("Cannot find the laptop with ID: %s", id)
	}
}

// type DBLaptopStore struct {
// }
