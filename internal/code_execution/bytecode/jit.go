package bytecode

import "fmt"

type FuncKind int

const (
	FuncJITCompiled           FuncKind = iota
	FuncPendingCompiledReturn          // waiting for return value to cache
	FuncCallingDynamic                 // may be compiled if all child functions become compiled
	FuncDynamic
)

type funcJITInfo struct {
	Kind            FuncKind
	CompiledAddress int
	CachedCalls     []*callInfo
}

type callInfo struct {
	args   []Value
	result Value
}

func (ci *callInfo) Equal(other callInfo) bool {
	if len(ci.args) != len(other.args) {
		return false
	}
	for i := range ci.args {
		if ci.args[i] != other.args[i] {
			return false
		}
	}
	return true
}

func findCallInfo(callInfos []*callInfo, target callInfo) (*callInfo, int) {
	for i, ci := range callInfos {
		if ci.Equal(target) {
			return ci, i
		}
	}
	return nil, 0
}

type FuncAddress int

type JITCompiler struct {
	bytecode      *Bytecode
	seenFunctions map[FuncAddress]*funcJITInfo
	pendingReturn map[FuncAddress][]*callInfo
}

func NewJITCompiler(bytecode *Bytecode) *JITCompiler {
	return &JITCompiler{
		bytecode:      bytecode,
		seenFunctions: make(map[FuncAddress]*funcJITInfo),
		pendingReturn: make(map[FuncAddress][]*callInfo),
	}
}

// GetCompiledAddress tries to compile the function at the given call instruction index and returns address of next instruction
// Returned address is either the compiled function address or the original function address if it cannot be compiled
func (jit *JITCompiler) GetCompiledAddress(callInstructionIndex int, args []Value) (FuncKind, int) {
	funcAddr := FuncAddress(getCallAddress(jit.bytecode.Instructions[callInstructionIndex]))
	currentCallInfo := callInfo{args: args}
	if funcAddr == -1 {
		return FuncDynamic, int(funcAddr) // error
	}

	if info, ok := jit.seenFunctions[funcAddr]; ok {
		switch info.Kind {
		case FuncDynamic:
			return FuncDynamic, int(funcAddr) // already known to be dynamic
		case FuncCallingDynamic:
			// worth checking again (maybe child functions are compiled now)
		case FuncPendingCompiledReturn:
			if inf, _ := findCallInfo(info.CachedCalls, currentCallInfo); inf != nil {
				return FuncPendingCompiledReturn, int(funcAddr) // Already has pending compilation for these arguments
			}
			jit.pendingReturn[funcAddr] = append(jit.pendingReturn[funcAddr], &currentCallInfo)
			return FuncDynamic, int(funcAddr)
		case FuncJITCompiled:
			for _, call := range info.CachedCalls {
				if call.Equal(currentCallInfo) {
					fmt.Printf("Using cached compiled function %s(%v) -> %v at address %d\n", jit.bytecode.FuncAddresses[int(funcAddr)].Name, call.args, call.result.Data, info.CompiledAddress)
					return FuncJITCompiled, info.CompiledAddress
				}
			}
			jit.pendingReturn[funcAddr] = append(jit.pendingReturn[funcAddr], &currentCallInfo)
			// is compiled but with different arguments
			return FuncDynamic, int(funcAddr)
		}
	}
	for _, arg := range args {
		if arg.Type == ValHeapPtr {
			jit.seenFunctions[funcAddr] = &funcJITInfo{Kind: FuncDynamic, CompiledAddress: int(funcAddr)}
			return FuncDynamic, int(funcAddr) // cannot JIT compile functions with heap arguments
		}
	}

	var farthestJump int
	// Check if function's instructions can be JIT compiled
	for i := int(funcAddr); i < len(jit.bytecode.Instructions); i++ {
		inst := jit.bytecode.Instructions[i]
		if inst.isJump() {
			farthestJump = max(farthestJump, inst.Operands[0])
		}
		if inst.isReturn() && i >= farthestJump {
			jit.pendingReturn[funcAddr] = append(jit.pendingReturn[funcAddr], &currentCallInfo)
			jit.seenFunctions[funcAddr] = &funcJITInfo{
				Kind:            FuncPendingCompiledReturn,
				CompiledAddress: int(funcAddr),
			}
			return FuncPendingCompiledReturn, int(funcAddr) // will compile after return value is known
		}

		if inst.Opcode == OpCall {
			compilable, kind := jit.checkCallCompilable(FuncAddress(inst.Operands[0]), funcAddr)
			if !compilable {
				return kind, int(funcAddr)
			}
			continue
		}

		if inst.hasSideEffects() {
			jit.seenFunctions[funcAddr] = &funcJITInfo{Kind: FuncDynamic, CompiledAddress: int(funcAddr)}
			return FuncDynamic, int(funcAddr) // has side effects, cannot JIT compile
		}
	}

	return FuncDynamic, int(funcAddr) // error (reached end of instructions without return)
}

func (jit *JITCompiler) NotifyReturn(funcAddrInt int, inputValues []Value, returnValue Value) {
	funcAddr := FuncAddress(funcAddrInt)
	if callInfos, ok := jit.pendingReturn[funcAddr]; ok {
		callInfo, idx := findCallInfo(callInfos, callInfo{args: inputValues})
		if callInfo == nil {
			return
		}
		callInfo.result = returnValue
		funcInfo := jit.seenFunctions[funcAddr]
		if funcInfo.Kind == FuncPendingCompiledReturn {
			funcInfo.CompiledAddress = jit.compile(len(callInfo.args), returnValue)
		}
		funcInfo.CachedCalls = append(funcInfo.CachedCalls, callInfo)
		fmt.Printf("Compiling function %s(%v) -> %v at address %d\n", jit.bytecode.FuncAddresses[funcAddrInt].Name, callInfo.args, callInfo.result.Data, funcAddr)
		callInfos = append(callInfos[:idx], callInfos[idx+1:]...)

		jit.pendingReturn[funcAddr] = callInfos
		if len(callInfos) == 0 {
			funcInfo.Kind = FuncJITCompiled
			delete(jit.pendingReturn, funcAddr)
		}
	}
}

func (jit *JITCompiler) checkCallCompilable(calledFuncAddr, funcAddr FuncAddress) (bool, FuncKind) {

	calledFuncInfo, ok := jit.seenFunctions[calledFuncAddr]
	if !ok {
		jit.seenFunctions[funcAddr] = &funcJITInfo{Kind: FuncCallingDynamic, CompiledAddress: int(funcAddr)}
		return false, FuncCallingDynamic
	}
	if calledFuncAddr == funcAddr {
		return true, 0 // recursive call, ignore
	}

	// called function is not compiled
	if calledFuncInfo.Kind != FuncJITCompiled {
		jit.seenFunctions[funcAddr] = &funcJITInfo{Kind: FuncCallingDynamic, CompiledAddress: int(funcAddr)}
		return false, FuncCallingDynamic
	}

	return true, 0 // might still be compilable (child function is compiled)
}

func (jit *JITCompiler) compile(numArgs int, returnValue Value) int {
	compiledAddr := len(jit.bytecode.Instructions)
	for i := range numArgs {
		jit.emit(OpStore, i)
	}
	if returnValue.Type != ValNil {
		jit.emit(OpConst, jit.addConstant(returnValue.Data))
		jit.emit(OpReturn)
		return compiledAddr
	}

	jit.emit(OpReturnVoid)
	return compiledAddr
}

func (jit *JITCompiler) emit(opcode byte, operands ...int) {
	jit.bytecode.Instructions = append(jit.bytecode.Instructions, Instruction{
		Opcode:   opcode,
		Operands: operands,
	})
}

func (jit *JITCompiler) replace(index int, opcode byte, operands ...int) {
	jit.bytecode.Instructions[index] = Instruction{
		Opcode:   opcode,
		Operands: operands,
	}
}

func (jit *JITCompiler) addConstant(value interface{}) int {
	for i, v := range jit.bytecode.Constants {
		if v == value {
			return i
		}
	}
	jit.bytecode.Constants = append(jit.bytecode.Constants, value)
	return len(jit.bytecode.Constants) - 1
}

func getCallAddress(inst Instruction) int {
	if inst.Opcode != OpCall {
		return -1
	}
	return inst.Operands[0]
}
