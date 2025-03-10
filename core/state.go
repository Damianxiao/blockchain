package core

import (
	"fmt"
	"log/slog"
)

// should set a interface

type state interface {
	put(data []byte, key string) error
	del(key string) error
	get(key string) error
}

//  todo account state

type contractState struct {
	//  todo  contract state , maybe one contract pair one data
	data map[string][]byte
}

func NewContractState() *contractState {
	return &contractState{
		data: make(map[string][]byte, 1024),
	}
}

func (s *contractState) put(key string, data []byte) error {
	s.data[key] = data
	return nil
}

func (s *contractState) del(key string) error {
	if _, ok := s.data[key]; !ok {
		slog.Info("contractState: delete key err", "err:", fmt.Errorf("data dont exist"))
	}
	delete(s.data, key)
	return nil
}

func (s *contractState) get(key string) ([]byte, error) {
	if _, ok := s.data[key]; !ok {
		slog.Info("contractState key err", "err:", fmt.Errorf("data dont exist"))
		return nil, fmt.Errorf("contractState:%v", "data dont exist")
	}
	return s.data[key], nil
}
