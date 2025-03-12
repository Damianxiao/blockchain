package core

import (
	"blockchain/pkg/utils/tool"
	"fmt"
)

type Instruction byte

const (
	instrPushInt  = 0x0a // push to stackt
	instrAdd      = 0x0b // 10 represent add
	instrPushByte = 0x0c // deal with char
	instrPack     = 0x0d // ?
	instrMinus    = 0x0e
	instrStore    = 0x0f // store to state
	instrGet      = 0x10
	instrMult     = 0x11
	instrDiv      = 0x12
)

type Stack struct {
	data []any
	sp   int
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]any, size),
		sp:   0,
	}
}

func (s *Stack) push(v any) {
	// s.data = append([]any{v}, s.data)
	// s.sp++
	s.data[s.sp] = v
	s.sp++
}

func (s *Stack) pop() any {
	s.sp--
	value := s.data[s.sp]
	s.data[s.sp] = nil
	// value := s.data[0]
	// slices.Delete(s.data, 1, 1)
	// s.sp--
	return value
}

type VM struct {
	stack         *Stack
	data          []byte
	ip            int // instruction pointer
	contractstate *contractState
}

func NewVM(data []byte, contractState *contractState) *VM {
	return &VM{
		stack:         NewStack(1024),
		data:          data,
		ip:            0,
		contractstate: contractState,
	}
}

func (vm *VM) run() error {
	for {
		instr := vm.data[vm.ip]
		err := vm.parseInstr(instr)
		vm.ip++

		if err != nil {
			return err
		}
		if vm.ip > len(vm.data)-1 {
			break
		}
	}
	return nil
}

func (vm *VM) parseInstr(instr byte) error {
	switch instr {
	case instrPushInt:
		value := int(vm.data[vm.ip-1])
		vm.stack.push(value)
	case instrAdd:
		add1 := vm.stack.pop().(int)
		add2 := vm.stack.pop().(int)
		res := add1 + add2
		vm.stack.push(res)
	case instrPushByte:
		b := byte(vm.data[vm.ip-1])
		vm.stack.push(b)
	case instrPack:
		n := vm.stack.pop().(int)
		str := make([]byte, n)
		for i := 0; i < n; i++ {
			str[i] = vm.stack.pop().(byte)
		}
		// push a []byte can be convert to stirng
		vm.stack.push(str)
	case instrMinus:
		add1 := vm.stack.pop().(int)
		add2 := vm.stack.pop().(int)
		res := add1 - add2
		vm.stack.push(res)
	case instrMult:
		add1 := vm.stack.pop().(int)
		add2 := vm.stack.pop().(int)
		res := add1 * add2
		vm.stack.push(res)
	case instrDiv:
		div1 := vm.stack.pop().(int)
		div2 := vm.stack.pop().(int)
		res := div1 / div2
		vm.stack.push(res)
	case instrStore:
		key := vm.stack.pop().([]byte)
		//  key should be data , but value is unknown
		data := vm.stack.pop()
		res := []byte{}
		switch v := data.(type) {
		case int:
			//  conv to 8 byte int
			//  why 8 byte? contract use same bit length to read data,if data is stirng, may not work
			res = tool.IntToBytes(int64(v))
		case []byte:
			res = data.([]byte)
		}
		// fmt.Printf("key: %v , value: %v", key, res)
		vm.contractstate.put(string(key), res)
	case instrGet:
		key := vm.stack.pop().([]byte)
		value, err := vm.contractstate.get(string(key))
		if err != nil {
			return fmt.Errorf("key not found")
		}
		vm.stack.push(value)
	}
	return nil
}
