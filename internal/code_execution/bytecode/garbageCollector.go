package bytecode

type GarbageCollector struct {
}

func (gc *GarbageCollector) Collect(heapPtr *[]Array, frames []Frame, frameIndex int) {
	mark := make(map[int]struct{})
	for i, frame := range frames {
		if i == frameIndex {
			continue
		}
		for _, local := range frame.locals {
			if local.Type != ValHeapPtr {
				continue
			}
			mark[local.Data.(int)] = struct{}{}
		}
	}
	heap2 := make([]Array, len(mark))
	for k, _ := range mark {
		heap2 = append(heap2, (*heapPtr)[k])
	}
	*heapPtr = heap2
}
