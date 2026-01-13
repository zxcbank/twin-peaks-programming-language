package bytecode

type GarbageCollector struct {
}

func (gc *GarbageCollector) Collect(heap []*Array, frames []Frame, frameIndex int) {
	markActiveArrPtrs := make(map[int]struct{})
	for i, frame := range frames {
		if i == frameIndex {
			continue
		}
		for _, local := range frame.locals {
			if local.Type != ValHeapPtr {
				continue
			}
			markActiveArrPtrs[local.Data.(int)] = struct{}{}
		}
	}

	for _, valueToDelete := range frames[frameIndex].locals {
		if valueToDelete.Type != ValHeapPtr {
			continue
		}
		if _, ok := markActiveArrPtrs[valueToDelete.Data.(int)]; !ok {
			heap[valueToDelete.Data.(int)] = nil
		}
	}
}
