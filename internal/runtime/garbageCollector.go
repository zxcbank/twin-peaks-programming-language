package runtime

type GarbageCollector struct {
}

func (gc *GarbageCollector) Collect(heap []*Array, frames []Frame, removedFrameIndex int) {
	markActivePtrs := make(map[int]struct{})
	for i, frame := range frames {
		if i == removedFrameIndex {
			continue
		}
		for _, local := range frame.locals {
			if local.Type != ValHeapPtr {
				continue
			}
			markActivePtrs[local.Data.(int)] = struct{}{}
		}
	}

	for _, valueToDelete := range frames[removedFrameIndex].locals {
		if valueToDelete.Type != ValHeapPtr {
			continue
		}
		ptr := valueToDelete.Data.(int)
		if _, ok := markActivePtrs[ptr]; !ok {
			heap[ptr] = nil
		}
	}
}
