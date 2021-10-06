package sample

import (
	"github.com/golang/protobuf/ptypes"
	"gitlab.com/keshavbhattad/pcbook/pb"
)

func NewKeyboard() *pb.Keyboard {
	keyboard := &pb.Keyboard{
		Layout:  randomKeyboardLayout(),
		Backlit: randomBool(),
	}
	return keyboard
}

func NewCPU() *pb.CPU {
	brand := randomCPUBrand()
	name := randomCPUName(brand)

	numberCores := randomInt(2, 8)
	numberThreads := randomInt(numberCores, 12)

	minGHz := randomFloat64(2.0, 3.5)
	maxGHz := randomFloat64(minGHz, 5.0)

	cpu := &pb.CPU{
		Brand:           brand,
		Name:            name,
		NumberOfCores:   uint32(numberCores),
		NumberOfThreads: uint32(numberThreads),
		MinGhz:          minGHz,
		MaxGhz:          maxGHz,
	}
	return cpu
}

func NewGPU() *pb.GPU {
	brand := randomGPUBrand()
	name := randomGPUName(brand)

	minGHz := randomFloat64(1.0, 1.5)
	maxGHz := randomFloat64(minGHz, 2.0)

	memory := &pb.Memory{
		Value: uint64(randomInt(2, 6)),
		Unit:  pb.Memory_GIGYBYTE,
	}

	gpu := &pb.GPU{
		Brand:  brand,
		Name:   name,
		MinGhz: minGHz,
		MaxGhz: maxGHz,
		Memory: memory,
	}
	return gpu
}

func NewRAM() *pb.Memory {
	ram := &pb.Memory{
		Value: uint64(randomInt(4, 6)),
		Unit:  pb.Memory_GIGYBYTE,
	}
	return ram
}

func NewSSD() *pb.Storage {
	memory := &pb.Memory{
		Value: uint64(randomInt(128, 1024)),
		Unit:  pb.Memory_GIGYBYTE,
	}

	ssd := &pb.Storage{
		Driver: pb.Storage_SSD,
		Memory: memory,
	}
	return ssd
}

func NewHDD() *pb.Storage {
	memory := &pb.Memory{
		Value: uint64(randomInt(1, 6)),
		Unit:  pb.Memory_TERABYTE,
	}

	hdd := &pb.Storage{
		Driver: pb.Storage_HDD,
		Memory: memory,
	}
	return hdd
}

func NewScreen() *pb.Screen {

	screen := &pb.Screen{
		ScreenSize: randomFloat32(13, 17),
		Resolution: randomScreenRresolution(),
		Panel:      randomScreenPanel(),
		Multitouch: randomBool(),
	}
	return screen
}

func NewLaptop() *pb.Laptop {
	brand := randomLaptopBrand()
	name := randomLaptopName(brand)
	laptop := &pb.Laptop{
		Id:       randomID(),
		Brand:    brand,
		Name:     name,
		Cpu:      NewCPU(),
		Ram:      NewRAM(),
		Screen:   NewScreen(),
		Keyboard: NewKeyboard(),
		Gpus:     []*pb.GPU{NewGPU()},
		Storages: []*pb.Storage{NewSSD(), NewHDD()},
		Weight: &pb.Laptop_WeightKg{
			WeightKg: randomFloat64(1.0, 3.0),
		},
		PriceInr:    randomFloat64(50000.0, 100000.0),
		ReleaseYear: uint32(randomInt(2012, 2021)),
		UpdatedAt:   ptypes.TimestampNow(),
	}
	return laptop
}

func RandomLaptopScore() float64 {
	return float64(randomInt(1, 10))
}
