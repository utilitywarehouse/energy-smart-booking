package inmemory

import (
	"bufio"
	"fmt"
	"os"
)

type MeterSerialNumberStore struct {
	meterSerialNumber map[string]struct{}
}

func NewMeterSerialNumber(filePath string) (*MeterSerialNumberStore, error) {
	msnMap, err := loadFromFile(filePath)
	if err != nil {
		return nil, err
	}

	return &MeterSerialNumberStore{
		meterSerialNumber: msnMap,
	}, nil
}

func (s *MeterSerialNumberStore) FindMeterSerialNumber(msn string) bool {
	_, ok := s.meterSerialNumber[msn]
	return ok
}

func loadFromFile(path string) (map[string]struct{}, error) {
	result := map[string]struct{}{}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open allowed meter serial number file, %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		result[line] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return result, nil
}
