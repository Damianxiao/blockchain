package core

import (
	"blockchain/pkg/utils/tool"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVMInt(t *testing.T) {
	data := []byte{0x01, 0x0a, 0x03, 0x0a, 0x0b}

	vm := NewVM(data, NewContractState())
	err := vm.run()
	// fmt.Println(vm.stack.data...)

	assert.Nil(t, err)
	res := vm.stack.pop()
	assert.Equal(t, 4, res.(int))
}

func TestStack(t *testing.T) {
	s := NewStack(1024)
	s.push(0x01)
	s.push(0x02)
	res := s.pop()
	assert.Equal(t, 0x02, res.(int))
}

func TestStringData(t *testing.T) {
	data := []byte{0x46, 0x0c, 0x79, 0x0c, 0x65, 0x0c, 0x6c, 0x0c, 0x6f, 0x0c, 0x20, 0x0c, 0x57, 0x0c, 0x6f, 0x0c, 0x72, 0x0c, 0x6c, 0x0c, 0x64, 0x0c, 0x21, 0x0c, 0x09, 0x0a, 0x0d}
	vm := NewVM(data, NewContractState())
	err := vm.run()
	// fmt.Println(vm.stack.pop().([]byte))
	assert.Nil(t, err)
}

func TestMinus(t *testing.T) {
	data := []byte{0x01, 0x0a, 0x03, 0x0a, 0x0e}
	vm := NewVM(data, NewContractState())
	err := vm.run()
	assert.Nil(t, err)
	res := vm.stack.pop().(int)
	assert.Equal(t, res, 2)
}

func TestInstrStoreString(t *testing.T) {
	// [f o o] 3 pack [f] 1 pack
	// [foo] [f] store
	// when value is string
	data := []byte{0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, 0x46, 0x0c, 0x01, 0x0a, 0x0d, 0x0f}

	vm := NewVM(data, NewContractState())
	err := vm.run()
	assert.Nil(t, err)
	assert.Equal(t, vm.contractstate.data["F"], []byte("OOF"))
}

func TestInstrStoreInt(t *testing.T) {
	data := []byte{0x01, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}

	vm := NewVM(data, NewContractState())
	err := vm.run()
	assert.Nil(t, err)
	value, err := vm.contractstate.get("OOF")
	assert.Nil(t, err)
	assert.Equal(t, tool.BytesToInt(value), int64(1))
}

func TestInstrMult(t *testing.T) {
	data := []byte{0x02, 0x0a, 0x05, 0x0a, 0x11}
	vm := NewVM(data, NewContractState())
	err := vm.run()
	assert.Nil(t, err)
	assert.Equal(t, vm.stack.data[0], int(10))
}

func TestInstrDiv(t *testing.T) {
	data := []byte{0x02, 0x0a, 0x04, 0x0a, 0x12}
	vm := NewVM(data, NewContractState())
	err := vm.run()
	assert.Nil(t, err)
	assert.Equal(t, vm.stack.data[0], int(2))
}

func TestInstrGet(t *testing.T) {
	store := []byte{0x01, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	key := []byte{0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, 0x10}
	key = append(store, key...)
	vm := NewVM(key, NewContractState())
	err := vm.run()
	assert.Nil(t, err)
	assert.Equal(t, vm.stack.data[0], tool.IntToBytes(1))
}
