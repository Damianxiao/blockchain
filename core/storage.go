package core

type Storage interface {
	Put(*Block) error
}

type MemoryStorage struct {
}

func NewStorage() *MemoryStorage {
	return &MemoryStorage{}
}

func (ms *MemoryStorage) Put(b *Block) error {
	return nil
}
