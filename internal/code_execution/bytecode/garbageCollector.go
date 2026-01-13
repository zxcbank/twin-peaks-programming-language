package bytecode

type GarbageCollector struct {
	heap *[]Array
}

func NewGarbageCollector(heap *[]Array) *GarbageCollector {
	return &GarbageCollector{heap: heap}
}

func (gc *GarbageCollector) Clear() {
	for i, array := range *gc.heap {
		if array.getReferenceCount() == 0 {
			*(gc.heap) = append((*(gc.heap))[:i], (*(gc.heap))[i+1:]...)
		}
	}
}

func (gc *GarbageCollector) OnReturnReferenceDecrement() {
	for _, array := range *gc.heap {
		array.decrementReference()
	}
}

func (gc *GarbageCollector) OnCallReferenceIncrement() {
	for _, array := range *gc.heap {
		array.incrementReference()
	}
}
