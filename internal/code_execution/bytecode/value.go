package bytecode

import "fmt"

// Value - значение в виртуальной машине
type Value struct {
	Type int
	Data interface{}
}

func (v Value) String() string {
	return fmt.Sprintf("%v", v.Data)
}

// Типы значений
const (
	ValInt = iota
	ValFloat
	ValString
	ValBool
	ValNil
	ValHeapPtr
)
