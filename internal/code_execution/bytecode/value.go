package bytecode

import "fmt"

// Value represents a value in vm during execution.
type Value struct {
	Type int
	Data interface{}
}

func (v Value) String() string {
	return fmt.Sprintf("%v", v.Data)
}

const (
	ValInt = iota
	ValFloat
	ValString
	ValBool
	ValNil
	ValHeapPtr
)
